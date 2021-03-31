package network

import (
	"errors"
	"fmt"
	"net"
	"time"
	"tinx/conf"
	"tinx/log"
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
	Name        string
	Type        string
	Ip          string
	Port        int
	NetWork     string
	readTimeout time.Duration
	Listener    net.Listener
	errorChan   chan error
	exitChan    chan struct{}
	manage      *ClientManage
	onConnect   func(c *TCPConnection)
	onClose     func(c *TCPConnection)
	route       *MsgHandle
}

func NewServer(config *conf.Zconfig) IServer {
	return &Server{
		Name:      config.Name,
		Ip:        config.Ip,
		NetWork:   config.NetWork,
		Port:      config.Port,
		Listener:  nil,
		Type:      config.Type,
		errorChan: make(chan error, 1),
		exitChan:  make(chan struct{}, 1),
		manage:    NewManage(config.MaxClients),
		route:     NewMsgHandler(uint32(config.PoolSize)),
	}
}

func (s *Server) Start() {
	addr, err := net.ResolveTCPAddr(s.NetWork, fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		s.errorChan <- err
		return
	}
	listener, err := net.ListenTCP(s.NetWork, addr)
	if err != nil {
		s.errorChan <- err
		return
	}
	s.Listener = listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("err:%s", err.Error())
			continue
		}
		if s.manage.Overload() {
			fmt.Println("overload")
			conn.Close()
			continue
		}
		// set keplive
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.SetKeepAlive(true)
			tc.SetKeepAlivePeriod(3 * time.Minute)
			tc.SetLinger(10)
		}
		c := NewTcpConnection(s, conn)
		if s.onConnect != nil {
			s.onConnect(c)
		}
		//add to manage
		s.manage.AddClient(c)
		go c.serveCon()
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
	if s.Listener == nil {
		return errors.New("server closed")
	}
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
