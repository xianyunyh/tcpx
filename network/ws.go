package network

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
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
