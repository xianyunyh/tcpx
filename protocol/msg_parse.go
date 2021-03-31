package protocol

import (
	"encoding/binary"
	"errors"
)


type MsgParser interface {
	Pack(msg Msger) []byte
	UnPack(raw []byte) (Msger, error)
}

type MsgParse struct {
	HeadLen uint32
}

func NewMsgParse(headLen uint32) *MsgParse {
	return &MsgParse{
		HeadLen: headLen,
	}
}

func (p *MsgParse) Pack(msg Msger) []byte {
	if msg == nil {
		return []byte{}
	}
	buf := make([]byte, int(p.HeadLen)+int(msg.GetLen()))
	binary.LittleEndian.PutUint32(buf[:p.HeadLen], msg.GetLen())
	copy(buf[p.HeadLen:], msg.GetData())
	return buf
}

func (p *MsgParse) UnPack(raw []byte) (Msger, error) {
	if len(raw) < int(p.HeadLen) {
		return nil, errors.New("to less raw")
	}
	lens := binary.LittleEndian.Uint32(raw[:4])
	data := make([]byte, lens)
	msg := NewMsg(data)
	msg.Len = lens
	return msg, nil
}
