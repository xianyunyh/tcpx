package network

import (
	"fmt"
	"io"
	"net"
	"time"
	"tinx/log"
)

type TCPConnection struct {
	server    *Server
	Conn      net.Conn
	Id        uint64
	Closed    bool
	msgChan   chan *Msg
	closeChan chan struct{}
}

func NewTcpConnection(con net.Conn) *TCPConnection {
	f, _ := con.(*net.TCPConn).File()
	return &TCPConnection{
		Conn:      con,
		Id:        uint64(f.Fd()),
		Closed:    false,
		closeChan: make(chan struct{}, 1),
		msgChan:   make(chan *Msg),
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

func (t *TCPConnection) Start() {
	t.server.manage.AddClient(t)
	go t.write()
	go t.read()
	log.Notice("%s", "connetion start")
}

func (t *TCPConnection) read() {
	if t.Closed {
		return
	}
	defer t.Close()
	log.Notice("%s", "connetion read begin")
	msgParse := NewMsgParse(4)
	for {
		head := make([]byte, 4)
		_, err := io.ReadFull(t.Conn, head)
		if err != nil {
			return
		}
		msg, err := msgParse.UnPack(head)
		if err != nil {
			log.Errorf("%s", err.Error())
			return
		}
		body := make([]byte, msg.Len)
		if _, err = io.ReadFull(t.Conn, body); err != nil {
			log.Errorf("%s", err.Error())
			return
		}
		msg.Data = body
		req := NewReuqest(t, msg)
		route := t.GetServer().route
		if route.WokerPoolSize > 0 {
			route.sendMsgToQueue(req)
		} else {
			go t.GetServer().route.Dispatch(req)
		}
	}
}

func (t *TCPConnection) write() {
	for {
		select {
		case <-t.closeChan:
			return
		case msg, ok := <-t.msgChan:
			if ok {
				fmt.Println(msg)
			}
		}
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
	//读完
	for len(t.msgChan) > 0 {
		time.Sleep(time.Microsecond)
	}
	close(t.msgChan)
	t.Conn.Close()
	t.closeChan <- struct{}{}
	close(t.closeChan)
	t.Closed = true
}
