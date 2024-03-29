package service

import (
	"FishCache/cache"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	STUnknown = iota
	STInited
	STRunning
	STStop
	// Heart beat
	MsgHeartbeat = iota

)

// SocketService struct
type SocketService struct {
	// hook
	onMessage    func(*Session, *Message)
	onConnect    func(*Session)
	onDisconnect func(*Session, error)

	sessions     *sync.Map
	hbInterval   time.Duration
	hbTimeout    time.Duration
	laddr        string
	status       int
	listener     net.Listener
	stopCh       chan error

	// local memory
	localCache  *cache.SCache

}

// Create a new socket service
func NewSocketService(addr string) (*SocketService, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &SocketService{
		sessions:   &sync.Map{},
		stopCh:     make(chan error),
		hbInterval: 0 * time.Second,
		hbTimeout:  0 * time.Second,
		laddr:      addr,
		status:     STInited,
		listener:   l,
	}
	return s, nil

}

// TODO : 使用hook在相应消息的过程中略显冗余
// RegMessageHandler register message handler
//func (s *SocketService) RegMessageHandler(handler func(*Session, *Message, *cache.SCache)) {
//	s.onMessage = handler
//}

// RegConnectHandler register connect handler
func (s *SocketService) RegConnectHandler(handler func(*Session)) {
	s.onConnect = handler
}

// RegDisconnectHandler register disconnect handler
func (s *SocketService) RegDisconnectHandler(handler func(*Session, error)) {
	s.onDisconnect = handler
}

func (s *SocketService) acceptHandler(ctx context.Context) {
	for {
		c, err := s.listener.Accept()
		if err != nil {
			s.stopCh <- err
			return
		}
		go s.connectHandler(ctx, c)
	}
}

func (s *SocketService) connectHandler (ctx context.Context, c net.Conn) {
	conn := NewConn(c, s.hbInterval, s.hbTimeout)
	session := NewSession(conn)

	s.sessions.Store(session.sID, session)

	connctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
		conn.Close()
		s.sessions.Delete(session.sID)
	}()

	go conn.read(connctx)
	go conn.write(connctx)

	if s.onConnect != nil {
		s.onConnect(session)
	}

	for {
		select {
		case err := <- conn.done:
			if s.onDisconnect != nil{
				s.onDisconnect(session, err)
			}
			return
		// Deal message
		case msg := <- conn.messageCh:
			commandString := string(msg.GetData())
			strings.Split(commandString, " ")

			}
		}
	}


// Serv Start socket service
func (s *SocketService) Serv() {

	s.status = STRunning
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		s.status = STStop
		cancel()
		_ = s.listener.Close()
	}()

	go s.acceptHandler(ctx)

	for {
		select {

		case <-s.stopCh:
			return
		}
	}
}

// Stop stop socket service with reason
func (s *SocketService) Stop(reason string) {
	s.stopCh <- errors.New(reason)
}


