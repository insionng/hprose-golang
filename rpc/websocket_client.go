/**********************************************************\
|                                                          |
|                          hprose                          |
|                                                          |
| Official WebSite: http://www.hprose.com/                 |
|                   http://www.hprose.org/                 |
|                                                          |
\**********************************************************/
/**********************************************************\
 *                                                        *
 * rpc/websocket_client.go                                *
 *                                                        *
 * hprose websocket client for Go.                        *
 *                                                        *
 * LastModified: Oct 8, 2016                              *
 * Author: Ma Bingyao <andot@hprose.com>                  *
 *                                                        *
\**********************************************************/

package rpc

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type reqeust struct {
	id   uint32
	data []byte
}

// WebSocketClient is hprose websocket client
type WebSocketClient struct {
	BaseClient
	limiter
	http.Header
	dialer    websocket.Dialer
	conn      *websocket.Conn
	nextid    uint32
	requests  chan reqeust
	responses map[uint32]chan socketResponse
	closed    bool
}

// NewWebSocketClient is the constructor of WebSocketClient
func NewWebSocketClient(uri ...string) (client *WebSocketClient) {
	client = new(WebSocketClient)
	client.initBaseClient()
	client.initLimiter()
	client.closed = false
	client.SetURIList(uri)
	client.SendAndReceive = client.sendAndReceive
	return
}

func newWebSocketClient(uri ...string) Client {
	return NewWebSocketClient(uri...)
}

func checkWebSocketAddresses(client *WebSocketClient, uriList []string) {
	for _, uri := range uriList {
		if u, err := url.Parse(uri); err == nil {
			if u.Scheme != "ws" && u.Scheme != "wss" {
				panic("This client desn't support " + u.Scheme + " scheme.")
			}
			if u.Scheme == "wss" {
				client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
			}
		}
	}
}

// SetURIList set a list of server addresses
func (client *WebSocketClient) SetURIList(uriList []string) {
	checkWebSocketAddresses(client, uriList)
	client.BaseClient.SetURIList(uriList)
}

func (client *WebSocketClient) close(err error) {
	client.cond.L.Lock()
	if err != nil && client.responses != nil {
		for _, response := range client.responses {
			response <- socketResponse{nil, err}
		}
	}
	client.responses = nil
	if client.conn != nil {
		client.conn.Close()
		client.conn = nil
	}
	client.reset()
	client.cond.L.Unlock()
}

// Close the client
func (client *WebSocketClient) Close() {
	client.closed = true
	client.close(errClientIsAlreadyClosed)
}

// TLSClientConfig returns the tls.Config in hprose client
func (client *WebSocketClient) TLSClientConfig() *tls.Config {
	return client.dialer.TLSClientConfig
}

// SetTLSClientConfig sets the tls.Config
func (client *WebSocketClient) SetTLSClientConfig(config *tls.Config) {
	client.dialer.TLSClientConfig = config
}

func (client *WebSocketClient) sendLoop() {
	conn := client.conn
	for request := range client.requests {
		err := conn.WriteMessage(websocket.BinaryMessage, request.data)
		if err != nil {
			client.close(err)
			break
		}
	}
	client.requests = nil
}

func (client *WebSocketClient) recvLoop() {
	conn := client.conn
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			client.close(err)
			break
		}
		if msgType == websocket.BinaryMessage {
			id := toUint32(data)
			client.cond.L.Lock()
			response := client.responses[id]
			if response != nil {
				response <- socketResponse{data[4:], nil}
				delete(client.responses, id)
			}
			client.unlimit()
			client.cond.L.Unlock()
		}
	}
	close(client.requests)
}

func (client *WebSocketClient) getConn(uri string) (err error) {
	if client.conn == nil {
		client.conn, _, err = client.dialer.Dial(uri, client.Header)
		if err != nil {
			return err
		}
		count := client.MaxConcurrentRequests
		client.requests = make(chan reqeust, count)
		client.responses = make(map[uint32]chan socketResponse, count)
		go client.sendLoop()
		go client.recvLoop()
	}
	return nil
}

func (client *WebSocketClient) sendAndReceive(
	data []byte, context *ClientContext) ([]byte, error) {
	id := atomic.AddUint32(&client.nextid, 1)
	buf := make([]byte, len(data)+4)
	fromUint32(buf, id)
	copy(buf[4:], data)
	response := make(chan socketResponse)
	client.cond.L.Lock()
	client.limit()
	if client.closed {
		client.cond.L.Unlock()
		return nil, errClientIsAlreadyClosed
	}
	if err := client.getConn(client.uri); err != nil {
		client.cond.L.Unlock()
		return nil, err
	}
	client.responses[id] = response
	client.cond.L.Unlock()
	client.requests <- reqeust{id, buf}
	select {
	case resp := <-response:
		return resp.data, resp.err
	case <-time.After(context.Timeout):
		client.cond.L.Lock()
		delete(client.responses, id)
		client.unlimit()
		client.cond.L.Unlock()
		return nil, ErrTimeout
	}
}
