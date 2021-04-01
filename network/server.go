package network

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"
	"tinx/conf"
	"tinx/log"
)

type Server struct {
	Name        string
	Type        string
	Ip          string
	Port        int
	NetWork     string
	Tlsconfig   *tls.Config
	listener    net.Listener
	readTimeout time.Duration
	errorChan   chan error
	exitChan    chan struct{}
	manage      *ClientManage
	onConnect   func(c *TCPConnection)
	onClose     func(c *TCPConnection)
	route       *MsgHandle
}

type OptionFunc func(s *Server)

func NewServer(config *conf.Zconfig) *Server {
	return &Server{
		Name:      config.Name,
		Ip:        config.Ip,
		NetWork:   config.NetWork,
		Port:      config.Port,
		listener:  nil,
		Type:      config.Type,
		errorChan: make(chan error, 1),
		exitChan:  make(chan struct{}, 1),
		manage:    NewManage(config.MaxClients),
		route:     NewMsgHandler(uint32(config.PoolSize)),
	}
}

func (s *Server) WithTlsOption(c *tls.Config) {
	s.Tlsconfig = c
}

func (s *Server) Start() {
	addr := fmt.Sprintf("%s:%d", s.Ip, s.Port)
	var err error
	var listener net.Listener
	if s.Tlsconfig != nil {
		listener, err = tls.Listen("tcp", addr, s.Tlsconfig)
	} else {
		listener, err = net.Listen(s.NetWork, addr)
	}
	if err != nil {
		s.errorChan <- err
		return
	}
	s.listener = listener
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
	if s.listener == nil {
		return errors.New("server closed")
	}
	s.manage.Clear()
	s.route.Close()
	return s.listener.Close()
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
