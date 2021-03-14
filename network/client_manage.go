package network

import "sync"

type ClientManage struct {
	clients map[uint64]*TCPConnection
	lock    sync.RWMutex
	max     uint32
}

func NewManage(max uint32) *ClientManage {
	return &ClientManage{
		lock:    sync.RWMutex{},
		clients: make(map[uint64]*TCPConnection),
		max:     max,
	}
}
func (c *ClientManage) Overload() bool {
	return c.Count() >= c.max
}
func (c *ClientManage) Count() uint32 {
	return uint32(len(c.clients))
}

func (c *ClientManage) AddClient(t *TCPConnection) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.clients[t.Id] = t
}

func (c *ClientManage) RemoveClient(t *TCPConnection) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.clients, t.Id)
}

func (c *ClientManage) Clear() {
	for _, v := range c.clients {
		v.Close()
	}
}
