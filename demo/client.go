// +build client
package main

import (
	"log"
	"net"
	"tinx/protocol"
)

func main() {
	con, err := net.Dial("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()
	p := protocol.NewMsgParse(4)

	for i := 0; i < 10; i++ {
		msg := protocol.NewMsg([]byte("hello world"))
		con.Write(p.Pack(msg))
	}

}
