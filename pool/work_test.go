package workpool

import "testing"

func TestPool(t *testing.T) {
	p := NewPool(2)
	p.Start()
	for i := 0; i < 50; i++ {
		msg := Msg{
			Id:   i,
			Data: []byte{byte(i)},
		}
		p.SendMsg(msg)
	}

}
