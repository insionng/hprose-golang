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
 * rpc/socket_service.go                                  *
 *                                                        *
 * hprose socket service for Go.                          *
 *                                                        *
 * LastModified: Oct 6, 2016                              *
 * Author: Ma Bingyao <andot@hprose.com>                  *
 *                                                        *
\**********************************************************/

package rpc

import (
	"bufio"
	"crypto/tls"
	"net"
	"reflect"
	"runtime"
	"sync"
)

// SocketContext is the hprose socket context for service
type SocketContext struct {
	serviceContext
	net.Conn
}

func (context *SocketContext) initSocketContext(
	service Service, conn net.Conn) {
	context.initServiceContext(service)
	context.Conn = conn
	return
}

func socketFixArguments(args []reflect.Value, context ServiceContext) {
	i := len(args) - 1
	switch args[i].Type() {
	case socketContextType:
		if c, ok := context.(*SocketContext); ok {
			args[i] = reflect.ValueOf(c)
		}
	case netConnType:
		if c, ok := context.(*SocketContext); ok {
			args[i] = reflect.ValueOf(c.Conn)
		}
	default:
		DefaultFixArguments(args, context)
	}
}

// SocketService is the hprose socket service
type SocketService struct {
	BaseService
	TLSConfig   *tls.Config
	contextPool chan *SocketContext
}

func (service *SocketService) initSocketService() {
	service.initBaseService()
	service.contextPool = make(chan *SocketContext, runtime.NumCPU()*8)
	service.FixArguments = socketFixArguments
	service.TLSConfig = nil
}

// ContextPoolSize returns the context pool size
func (service *SocketService) ContextPoolSize() int {
	return cap(service.contextPool)
}

// SetContextPoolSize sets the context pool size
func (service *SocketService) SetContextPoolSize(value int) {
	service.contextPool = make(chan *SocketContext, value)
}

func (service *SocketService) acquireContext() (context *SocketContext) {
	select {
	case context = <-service.contextPool:
		return
	default:
		return new(SocketContext)
	}
}

func (service *SocketService) releaseContext(context *SocketContext) {
	select {
	case service.contextPool <- context:
	default:
	}
}

func (service *SocketService) serveConn(conn net.Conn) {
	context := new(SocketContext)
	context.initSocketContext(service, conn)
	event := service.Event
	defer func() {
		if e := recover(); e != nil {
			err := NewPanicError(e)
			fireErrorEvent(event, err, context)
		}
	}()
	if err := fireAcceptEvent(event, context); err != nil {
		fireErrorEvent(event, err, context)
		return
	}
	handler := new(connHandler)
	handler.conn = conn
	handler.serve(service)
	if err := fireCloseEvent(event, context); err != nil {
		fireErrorEvent(event, err, context)
	}
}

type acceptEvent interface {
	OnAccept(context *SocketContext)
}

type acceptEvent2 interface {
	OnAccept(context *SocketContext) error
}

type closeEvent interface {
	OnClose(context *SocketContext)
}

type closeEvent2 interface {
	OnClose(context *SocketContext) error
}

func fireAcceptEvent(event ServiceEvent, context *SocketContext) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = NewPanicError(e)
		}
	}()
	switch event := event.(type) {
	case acceptEvent:
		event.OnAccept(context)
	case acceptEvent2:
		err = event.OnAccept(context)
	}
	return err
}

func fireCloseEvent(event ServiceEvent, context *SocketContext) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = NewPanicError(e)
		}
	}()
	switch event := event.(type) {
	case closeEvent:
		event.OnClose(context)
	case closeEvent2:
		err = event.OnClose(context)
	}
	return err
}

type connHandler struct {
	sync.Mutex
	conn net.Conn
}

func (handler *connHandler) serve(service *SocketService) {
	reader := bufio.NewReader(handler.conn)
	var data packet
	for {
		if err := recvData(reader, &data); err != nil {
			break
		}
		if data.fullDuplex {
			go handler.handle(service, data)
		} else {
			handler.handle(service, data)
		}
	}
	handler.conn.Close()
}

func (handler *connHandler) handle(service *SocketService, data packet) {
	context := service.acquireContext()
	context.initSocketContext(service, handler.conn)
	data.body = service.Handle(data.body, context)
	if data.fullDuplex {
		handler.Lock()
	}
	err := sendData(handler.conn, data)
	if data.fullDuplex {
		handler.Unlock()
	}
	if err != nil {
		fireErrorEvent(service.Event, err, context)
	}
	service.releaseContext(context)
}
