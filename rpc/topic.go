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
 * rpc/topic.go                                           *
 *                                                        *
 * hprose push topic for Go.                              *
 *                                                        *
 * LastModified: Sep 13, 2016                             *
 * Author: Ma Bingyao <andot@hprose.com>                  *
 *                                                        *
\**********************************************************/

package rpc

import "time"

type message struct {
	Detector chan bool
	Result   interface{}
}

type topic struct {
	*time.Timer
	Request   interface{}
	Messages  chan *message
	Count     int64
	Heartbeat time.Duration
}

func newTopic(heartbeat time.Duration) *topic {
	t := new(topic)
	t.Heartbeat = heartbeat
	return t
}
