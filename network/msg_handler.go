package network

import (
	"tinx/log"
)

type MsgHandle struct {
	handlers      map[uint32]Handler
	WokerPoolSize uint32
	TaskQueue     []chan *Request
}

func NewMsgHandler(size uint32) *MsgHandle {
	return &MsgHandle{
		handlers:      make(map[uint32]Handler),
		WokerPoolSize: size,
		TaskQueue:     make([]chan *Request, size),
	}
}

func (m *MsgHandle) Dispatch(req *Request) {
	msgId := req.GetMsgId()
	h, ok := m.handlers[msgId]
	if !ok {
		log.Errorf("%s", "handler not register")
		return
	}
	h.BeforeRequest(req)
	h.DoRequest(req)
	h.AfterRequest(req)
}
func (m *MsgHandle) newWork(i int, reqChan chan *Request) {
	for {
		select {
		case req := <-reqChan:
			m.Dispatch(req)
		}
	}
}

func (m *MsgHandle) sendMsgToQueue(req *Request) {
	workerID := req.conn.Id % uint64(m.WokerPoolSize)
	m.TaskQueue[workerID] <- req
}

func (m *MsgHandle) startWorkPool() {
	for i := 0; i < int(m.WokerPoolSize); i++ {
		m.TaskQueue[i] = make(chan *Request, m.WokerPoolSize)
		go m.newWork(i, m.TaskQueue[i])
	}
}

func (m *MsgHandle) Close() {
	for _, v := range m.TaskQueue {
		close(v)
	}
}

type Handler interface {
	BeforeRequest(req *Request)
	DoRequest(req *Request)
	AfterRequest(req *Request)
}
