package protocol

import (
	"io"
	"net"
	"tinx/log"
)

var msgReaders = make(map[string]MessageReader)

func init() {
	msgReaders["tcp"] = newTcpReader()
	msgReaders["websocket"] = newWsReader()
}

func GetMsgReader(name string, conn net.Conn) MessageReader {
	if r, ok := msgReaders[name]; ok {
		r.SetConnetion(conn)
		return r
	}
	return nil
}
func RegisterReader(name string, r MessageReader) {
	msgReaders[name] = r
}

func newTcpReader() MessageReader {
	return &RawMessage{
		parser: NewMsgParse(4),
	}
}

func newWsReader() MessageReader {
	return &WebsocketMessage{
		parser: NewWsMsgParse(4),
	}
}

type MessageReader interface {
	ReadData(buf io.Reader) (Msger, error)
	SetConnetion(con net.Conn)
}

type RawMessage struct {
	parser MsgParser
	conn   net.Conn
}

func (r *RawMessage) SetConnetion(con net.Conn) {
	r.conn = con
}
func (r *RawMessage) ReadData(buf io.Reader) (msg Msger, err error) {
	head := make([]byte, 4)
	_, err = io.ReadFull(buf, head)
	if err != nil {
		return nil, err
	}
	msg, err = r.parser.UnPack(head)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}
	body := make([]byte, msg.GetLen())
	if _, err = io.ReadFull(buf, body); err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}
	msg.SetData(body)
	return msg, nil
}
