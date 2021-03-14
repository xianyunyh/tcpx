package network

type Request struct {
	conn *TCPConnection
	msg  Msger
}

func NewReuqest(con *TCPConnection, data Msger) *Request {
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
func (r *Request) setData(data Msger) {
	r.msg = data
}

func (r *Request) GetConnection() *TCPConnection {
	return r.conn
}
func (r *Request) GetData() Msger {
	return r.msg
}
