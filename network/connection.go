package network

import (
	"bufio"
	"io"
	"net"
	"runtime"
	"tinx/log"
	"tinx/protocol"
)

type TCPConnection struct {
	server    *Server
	Conn      net.Conn
	Id        uint64
	Closed    bool
	reader    protocol.MessageReader
	closeChan chan struct{}
}

func NewTcpConnection(s *Server, con net.Conn) *TCPConnection {
	f, _ := con.(*net.TCPConn).File()
	return &TCPConnection{
		server:    s,
		Conn:      con,
		Id:        uint64(f.Fd()),
		Closed:    false,
		closeChan: make(chan struct{}, 1),
	}
}

func (t *TCPConnection) GetTcpConnection() net.Conn {
	return t.Conn
}

func (t *TCPConnection) GetServer() *Server {
	return t.server
}

func (t *TCPConnection) GetManage() *ClientManage {
	return t.server.manage
}

func (t *TCPConnection) serveCon() {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			log.Errorf("serving %s panic error: %s, stack:\n %s", t.Conn.RemoteAddr(), err, string(buf))
		}
		t.server.manage.RemoveClient(t)
		t.Close()
	}()
	log.Notice("%s", "connetion start")
	t.read()
}
func (t *TCPConnection) read() {
	if t.Closed {
		return
	}
	defer t.Close()
	buf := bufio.NewReader(t.Conn)
	reader := protocol.GetMsgReader(t.server.Type, t.Conn)
	if reader == nil {
		log.Errorf("%s not register messge reader")
		return
	}
	for {
		msg, err := reader.ReadData(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("read error %s", err.Error())
			return
		}
		req := NewReuqest(t, msg)
		t.dispatch(req)
	}
}

func (t *TCPConnection) dispatch(req *Request) {
	route := t.GetServer().route
	if route.WokerPoolSize > 0 {
		route.sendMsgToQueue(req)
	} else {
		go func() {
			t.GetServer().route.Dispatch(req)
		}()
	}
}
func (t *TCPConnection) Close() {
	if t.Closed {
		return
	}
	if t.GetServer().onClose != nil {
		on := t.GetServer().onClose
		on(t)
	}
	t.server.manage.RemoveClient(t)
	t.Conn.Close()
	t.closeChan <- struct{}{}
	close(t.closeChan)
	t.Closed = true
}
