package network

import (
	"encoding/binary"
	"errors"
)

type Msger interface {
	GetLen() uint32
	GetData() []byte
}
type Msg struct {
	Len  uint32
	Data []byte
}

func NewMsg(data []byte) *Msg {
	return &Msg{
		Len:  uint32(len(data)),
		Data: data,
	}
}
func (m *Msg) GetLen() uint32 {
	return m.Len
}
func (m *Msg) GetData() []byte {
	if m.Data == nil {
		return []byte{}
	}
	return m.Data
}

type MsgParse struct {
	HeadLen uint32
}

func NewMsgParse(headLen uint32) *MsgParse {
	return &MsgParse{
		HeadLen: headLen,
	}
}

func (p *MsgParse) Pack(msg *Msg) []byte {
	if msg == nil {
		return []byte{}
	}
	buf := make([]byte, int(p.HeadLen)+int(msg.GetLen()))
	binary.LittleEndian.PutUint32(buf[:p.HeadLen], msg.GetLen())
	copy(buf[p.HeadLen:], msg.GetData())
	return buf
}

func (p *MsgParse) UnPack(raw []byte) (*Msg, error) {
	if len(raw) < int(p.HeadLen) {
		return nil, errors.New("to less raw")
	}
	lens := binary.LittleEndian.Uint32(raw[:4])
	data := make([]byte, lens)
	msg := NewMsg(data)
	msg.Len = lens
	return msg, nil
}
