package workpool

import (
	"context"
	"errors"
	"fmt"
)

type Msg struct {
	Id   int
	Data []byte
}

type Runner interface {
	Run(msg Msg)
}
type WorkerPool struct {
	Size      int
	Closed    bool
	closeChan chan struct{}
	Msgs      chan Msg
	runner    Runner
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewPool(size int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		Size:      size,
		Closed:    false,
		closeChan: make(chan struct{}),
		Msgs:      make(chan Msg),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (p *WorkerPool) SetRunner(r Runner) {
	p.runner = r
}

func (p *WorkerPool) Start() {
	if p.Size <= 0 {
		p.Size = 1
	}

	for i := 0; i < p.Size; i++ {
		go p.doWork(i)
	}
}

func (p *WorkerPool) SendMsg(msg Msg) error {
	if p.Closed {
		return errors.New("pool is closed")
	}
	p.Msgs <- msg
	return nil
}

func (p *WorkerPool) doWork(id int) {
	for {
		select {
		case msg := <-p.Msgs:
			fmt.Println(msg.Data)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *WorkerPool) Stop() {
	if p.Closed {
		return
	}
	p.Closed = true
	p.cancel()
}
