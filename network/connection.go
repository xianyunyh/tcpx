package network

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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

func NewTcpConnection(s *Server, con net.Conn) *TCPConnection {
	f, _ := con.(*net.TCPConn).File()
	return &TCPConnection{
		server:    s,
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
	go t.read()
	log.Notice("%s", "connetion start")
}

func (t *TCPConnection) handleShake() error {
	buf := bufio.NewReader(t.Conn)
	request, err := http.ReadRequest(buf)
	if err != nil {
		return err
	}
	connetion := request.Header.Get("Connection")
	upgrade := request.Header.Get("Upgrade")
	webSocketKey := request.Header.Get("Sec-WebSocket-Key")
	webScoketVersion := request.Header.Get("Sec-WebSocket-Version")
	if connetion != "Upgrade" {
		return errors.New("upgrade")
	}
	if upgrade != "websocket" {
		return errors.New("websocket")
	}
	if webScoketVersion != "13" {
		return errors.New("verison ")
	}

	p := bytes.NewBuffer(nil)
	p.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	p.WriteString("Upgrade: websocket\r\n")
	p.WriteString("Connection: Upgrade\r\n")
	p.WriteString("Sec-WebSocket-Accept:" + computeAcceptKey(webSocketKey) + "\r\n")
	p.WriteString("\r\n")
	_, err = t.Conn.Write(p.Bytes())
	if err != nil {
		return err
	}
	return nil
}
func (t *TCPConnection) read() {
	if t.Closed {
		return
	}
	defer t.Close()
	if t.server.Type == "ws" {
		if err := t.handleShake(); err != nil {
			log.Errorf("handle error %s\n", err.Error())
			return
		}
	}
	for {
		var msg Msger
		var err error
		if t.server.Type == "ws" {
			msg, err = t.readWsData()
		} else {
			msg, err = t.readTcpData()

		}
		if err != nil {
			return
		}
		req := NewReuqest(t, msg)
		// return
		route := t.GetServer().route
		if route.WokerPoolSize > 0 {
			route.sendMsgToQueue(req)
		} else {
			go t.GetServer().route.Dispatch(req)
		}
	}
}

func (t *TCPConnection) readTcpData() (msg Msger, err error) {
	msgParse := t.server.parser
	head := make([]byte, 4)
	_, err = io.ReadFull(t.Conn, head)
	if err != nil {
		return nil, err
	}
	msg, err = msgParse.UnPack(head)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}
	body := make([]byte, msg.GetLen())
	if _, err = io.ReadFull(t.Conn, body); err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}
	msg.SetData(body)
	return msg, nil
}

func (t *TCPConnection) readWsData() (Msger, error) {
	//1. 读取头部
	msg := NewMsg(nil)
	var err error
	first := make([]byte, 2)
	_, err = io.ReadFull(t.Conn, first)
	if err != nil {
		return nil, err
	}
	isLast := (first[0] & 0x80) > 0
	if !isLast {
		fmt.Println("not last")
	}
	//opcode
	if first[0] != 0x81 {
		fmt.Printf(" first =%x \n", first[0])
	}

	useMask := (first[1] & 0x80) > 0
	masks := []byte{}

	//2. 读取len
	lens := first[1] & 0x7f // 01111111
	var l uint64
	l = uint64(uint8(lens))
	if lens >= 126 {
		if lens^0x7e == 0 {
			temp := make([]byte, 2)
			io.ReadFull(t.Conn, temp)
			t := binary.LittleEndian.Uint16(temp)
			l = uint64(t)
		} else {
			temp := make([]byte, 8)
			io.ReadFull(t.Conn, temp)
			l = binary.LittleEndian.Uint64(temp)
		}
	}
	// 3. 读取mask
	if useMask {
		masks = make([]byte, 4)
		io.ReadFull(t.Conn, masks)
	}
	//4. 读取payload
	payload := make([]byte, l)
	_, err = io.ReadFull(t.Conn, payload)
	if err != nil {
		return nil, err
	}

	if useMask {
		for i, v := range payload {
			j := i
			if i >= len(masks) {
				j = (i - len(masks)) % len(masks)
			}
			payload[i] = masks[j] ^ v
		}
	}
	msg.SetData(payload)
	return msg, nil
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
