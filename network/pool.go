package network

import (
	"sync"
	"tinx/protocol"
)

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
}

func GetRequest(conn *TCPConnection, msg protocol.Msger) *Request {
	req := requestPool.Get().(*Request)
	req.conn = conn
	req.msg = msg
	return req
}
func FreeRequest(req *Request) {
	req.conn = nil
	req.msg = nil
	requestPool.Put(req)
}
