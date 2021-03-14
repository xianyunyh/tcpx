// +build
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"tinx/network"
)

func main() {
	con, err := net.Dial("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()
	p := network.NewMsgParse(4)

	for i := 0; i < 10; i++ {
		msg := network.NewMsg([]byte("hello world"))
		con.Write(p.Pack(msg))
	}
	timer := time.NewTicker(10 * time.Second)
	go func(con net.Conn) {
		defer con.Close()
		for {
			head := make([]byte, p.HeadLen)
			_, err := io.ReadFull(con, head)
			if err != nil {
				return
			}
			msg, err := p.UnPack(head)
			body := make([]byte, msg.GetLen())
			_, err = io.ReadFull(con, body)
			if err != nil {
				return
			}
			fmt.Println(string(body))
		}
	}(con)
	select {
	case <-timer.C:
		con.Close()
		return
	}
}
