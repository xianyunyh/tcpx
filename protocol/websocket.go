package protocol

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

const (
	TextType    = 0x81
	BinnaryText = 0x82
	maskLen     = 4
)

type WsMsgParse struct {
	HeadLen uint32
}

func NewWsMsgParse(headLen uint32) *WsMsgParse {
	return &WsMsgParse{
		HeadLen: headLen,
	}
}

func (p *WsMsgParse) Pack(msg Msger) []byte {
	if msg == nil {
		return []byte{}
	}
	data := msg.GetData()
	lens := msg.GetLen()

	buf := bytes.NewBuffer(nil)
	buf.Write([]byte{0x81})
	if lens < 126 {
		buf.Write([]byte{byte(lens)})
	} else if lens >= 126 && lens < 655535 {
		temp := make([]byte, 3)
		temp[0] = 126
		binary.LittleEndian.PutUint16(temp[1:], uint16(lens))
		buf.Write(temp)
	} else {
		temp := make([]byte, 1+8)
		temp[0] = 127
		binary.LittleEndian.PutUint64(temp[1:], uint64(lens))
	}
	//Server Do Not Use Mask
	buf.Write(data)
	return buf.Bytes()
}

func (p *WsMsgParse) randMsks() []byte {
	//todo
	return []byte{0x1, 0x2, 0x3, 0x4}
}

func (p *WsMsgParse) UnPack(raw []byte) (Msger, error) {
	buf := bytes.NewBuffer(raw)
	//max 2+ 8
	// 1. 读取首字节
	first := raw[0]
	_ = (first & 0x80) > 0
	second := raw[1]
	//2. 读取len
	lens := second & 0x7f // 01111111
	var l uint64
	l = uint64(uint8(lens))
	if lens >= 126 {
		if lens^0x7e == 0 {
			temp := make([]byte, 2)
			buf.Read(temp)
			t := binary.BigEndian.Uint16(temp)
			l = uint64(t)
		} else {
			temp := make([]byte, 8)
			buf.Read(temp)
			l = binary.BigEndian.Uint64(temp)
		}
	}
	useMask := (first & 0x80) > 0
	masks := []byte{}
	if useMask {
		masks = make([]byte, 4)
		buf.Read(masks)
	}
	data := make([]byte, lens)
	msg := NewMsg(data)
	msg.Len = uint32(l)
	return msg, nil
}

func computeAcceptKey(challengeKey string) string {
	h := sha1.New()
	h.Write([]byte(challengeKey))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type WebsocketMessage struct {
	parser MsgParser
	conn   net.Conn
}

//协议握手
func (r *WebsocketMessage) handleShake(buf *bufio.Reader) error {
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
	_, err = r.conn.Write(p.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (r *WebsocketMessage) SetConnetion(con net.Conn) {
	r.conn = con
}

//读取数据
func (r *WebsocketMessage) ReadData(buf io.Reader) (Msger, error) {

	if buf == nil {
		return nil, errors.New("buf is nil")
	}
	//握手
	err := r.handleShake(buf.(*bufio.Reader))
	if err != nil {
		return nil, err
	}
	//1. 读取头部
	first := make([]byte, 2)
	_, err = io.ReadFull(buf, first)
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
	lens := first[1] & 0x7f
	var l uint64
	l = uint64(uint8(lens))
	if lens >= 126 {
		if lens^0x7e == 0 {
			temp := make([]byte, 2)
			io.ReadFull(buf, temp)
			t := binary.LittleEndian.Uint16(temp)
			l = uint64(t)
		} else {
			temp := make([]byte, 8)
			io.ReadFull(buf, temp)
			l = binary.LittleEndian.Uint64(temp)
		}
	}
	// 3. 读取mask
	if useMask {
		masks = make([]byte, 4)
		io.ReadFull(buf, masks)
	}
	//4. 读取payload
	payload := make([]byte, l)
	_, err = io.ReadFull(buf, payload)
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
	msg := NewMsg(nil)
	msg.SetData(payload)
	return msg, nil
}
