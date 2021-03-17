package network

import (
	"fmt"
	"log"
	"net"
	"tinx/conf"
)

type IServer interface {
	Start()
	Stop() error
	Serve()
	SetOnClose(func(c *TCPConnection))
	SetOnConnect(func(c *TCPConnection))
	AddHandler(id uint32, handle Handler)
}

type Server struct {
	Name      string
	IpVer     string
	Ip        string
	Port      int
	Listener  net.Listener
	errorChan chan error
	exitChan  chan struct{}
	manage    *ClientManage
	onConnect func(c *TCPConnection)
	onClose   func(c *TCPConnection)
	route     *MsgHandle
}

func NewServer(config *conf.Zconfig) IServer {
	return &Server{
		Name:      config.Name,
		Ip:        config.Ip,
		IpVer:     config.IpVer,
		Port:      config.Port,
		Listener:  nil,
		errorChan: make(chan error, 1),
		exitChan:  make(chan struct{}, 1),
		manage:    NewManage(config.MaxClients),
		route:     NewMsgHandler(uint32(config.PoolSize)),
	}
}
func (s *Server) Start() {
	addr, err := net.ResolveTCPAddr(s.IpVer, fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		s.errorChan <- err
		return
	}
	listener, err := net.ListenTCP(s.IpVer, addr)
	if err != nil {
		s.errorChan <- err
		return
	}
	s.Listener = listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if s.manage.Overload() {
			fmt.Println("overload")
			conn.Close()
			continue
		}
		c := NewTcpConnection(s, conn)
		if s.onConnect != nil {
			s.onConnect(c)
		}
		c.Start()
	}
}

func (s *Server) GetClientManage() *ClientManage {
	return s.manage
}

func (s *Server) SetOnConnect(callback func(c *TCPConnection)) {
	s.onConnect = callback
}
func (s *Server) SetOnClose(callback func(c *TCPConnection)) {
	s.onClose = callback
}

func (s *Server) AddHandler(id uint32, handle Handler) {
	s.route.handlers[id] = handle
}

func (s *Server) Stop() error {
	s.manage.Clear()
	s.route.Close()
	return s.Listener.Close()
}

func (s *Server) Serve() {
	go func() {
		s.Start()
		s.exitChan <- struct{}{}
	}()
	s.route.startWorkPool()
	select {
	case err := <-s.errorChan:
		s.Stop()
		fmt.Println(err.Error())
	case <-s.exitChan:
		s.Stop()
	}
}
