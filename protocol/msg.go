package protocol

type Msger interface {
	GetLen() uint32
	SetLen(l uint32)
	GetData() []byte
	SetData(data []byte)
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
func (m *Msg) SetLen(l uint32) {
	m.Len = l
}
func (m *Msg) GetData() []byte {
	if m.Data == nil {
		return []byte{}
	}
	return m.Data
}

func (m *Msg) SetData(data []byte) {
	m.Data = data
}
