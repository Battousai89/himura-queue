package server

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"himura-queue/internal/protocol"
	"himura-queue/internal/worker"
)

type Handler func(*protocol.Frame) *protocol.Frame

type TCPServer struct {
	listener net.Listener
	pool     *worker.Pool
	handler  Handler
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

func NewTCPServer(addr string, handler Handler, pool *worker.Pool) (*TCPServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPServer{
		listener: listener,
		pool:     pool,
		handler:  handler,
	}, nil
}

func (s *TCPServer) Start(ctx context.Context) {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		s.Stop()
	}()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			log.Printf("accept error: %v", err)
			continue
		}
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	for {
		frame, err := protocol.DecodeFrame(conn)
		if err != nil {
			if err != io.EOF && err != protocol.ErrEOF {
				log.Printf("decode error: %v", err)
			}
			return
		}
		s.pool.Submit(func() {
			resp := s.handler(frame)
			if resp != nil {
				conn.Write(protocol.EncodeFrame(resp))
			}
		})
	}
}

func (s *TCPServer) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()
	s.listener.Close()
	s.wg.Wait()
}
