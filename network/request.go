package network

import (
	"tinx/protocol"
)

type Request struct {
	conn *TCPConnection
	msg  protocol.Msger
}

func NewReuqest(con *TCPConnection, data protocol.Msger) *Request {
	return &Request{
		conn: con,
		msg:  data,
	}
}

func (r *Request) GetMsgId() uint32 {
	return 0
}
func (r *Request) setConnection(conn *TCPConnection) {
	r.conn = conn
}
func (r *Request) setMsg(n protocol.Msger) {
	r.msg = n
}

func (r *Request) GetConnection() *TCPConnection {
	return r.conn
}

func (r *Request) GetMsg() protocol.Msger {
	return r.msg
}
