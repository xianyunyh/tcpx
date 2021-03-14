package main

import (
	"fmt"
	"log"
	"tinx/conf"
	"tinx/network"
)

type PingHandle struct {
}

func (p *PingHandle) BeforeRequest(req *network.Request) {

	log.Println("before")
}
func (p *PingHandle) DoRequest(req *network.Request) {
	m := network.NewMsgParse(4)
	msg := network.NewMsg([]byte("hello world"))
	req.GetConnection().GetTcpConnection().Write([]byte(m.Pack(msg)))
}
func (p *PingHandle) AfterRequest(req *network.Request) {
	log.Println("after")
}

func main() {

	conf := &conf.Zconfig{
		Name:       "test",
		Ip:         "127.0.0.1",
		Port:       9090,
		IpVer:      "tcp",
		MaxClients: 100,
		PoolSize:   10,
	}
	server := network.NewServer(conf)
	//回调
	server.SetOnConnect(func(c *network.TCPConnection) {
		fmt.Println("client connect")
	})
	server.SetOnClose(func(c *network.TCPConnection) {
		fmt.Println("client closed")
	})
	ping := &PingHandle{}
	server.AddHandler(0, ping)
	server.Serve()

}
