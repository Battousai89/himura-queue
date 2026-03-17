package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"himura-queue/internal/deduplication"
	"himura-queue/internal/persistence"
	"himura-queue/internal/protocol"
	"himura-queue/internal/queue"
	"himura-queue/internal/worker"
)

type Config struct {
	TCPPort          int
	HTTPPort         int
	MinWorkers       int
	MaxWorkers       int
	ShardCount       int
	SnapshotPath     string
	SnapshotInterval time.Duration
}

type Server struct {
	config       Config
	tcpServer    *TCPServer
	httpServer   *http.Server
	queueManager *queue.Manager
	dedup        *deduplication.Deduplicator
	snapshotter  *persistence.Snapshotter
	pool         *worker.Pool
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewServer(config Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pool := worker.NewPool(1024, config.MinWorkers, config.MaxWorkers, 30*time.Second)
	qm := queue.NewManager(config.ShardCount)
	dedup := deduplication.NewDeduplicator(5*time.Minute, "data/ack.bin")
	snapshotter := persistence.NewSnapshotter(config.SnapshotPath, config.SnapshotInterval)

	lastAck := dedup.GetLastAck()
	qm.SetLastAck(lastAck)

	messages, err := snapshotter.Load()
	if err != nil {
		log.Printf("Failed to load snapshot: %v", err)
	} else {
		log.Printf("Loaded %d messages from snapshot", len(messages))
	}
	for _, msg := range messages {
		if msg.ID > lastAck {
			qm.Push(msg.Queue, msg.Payload, msg.Priority, msg.Delay)
		}
	}

	s := &Server{
		config:       config,
		queueManager: qm,
		dedup:        dedup,
		snapshotter:  snapshotter,
		pool:         pool,
		ctx:          ctx,
		cancel:       cancel,
	}

	s.tcpServer, _ = NewTCPServer(fmt.Sprintf(":%d", config.TCPPort), s.handleFrame, pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: mux,
	}

	return s, nil
}

func (s *Server) handleFrame(frame *protocol.Frame) *protocol.Frame {
	switch frame.Command {
	case protocol.CmdPush:
		return s.handlePush(frame.Data)
	case protocol.CmdPop:
		return s.handlePop(frame.Data)
	case protocol.CmdAck:
		return s.handleAck(frame.Data)
	case protocol.CmdStatus:
		return s.handleStatus(frame.Data)
	default:
		return nil
	}
}

func (s *Server) handlePush(data []byte) *protocol.Frame {
	req, err := protocol.DecodePushRequest(data)
	if err != nil {
		return nil
	}

	id := s.queueManager.Push(req.Queue, req.Payload, req.Priority, time.Duration(req.Delay))
	resp := protocol.EncodePushResponse(&protocol.PushResponse{ID: id})
	return &protocol.Frame{Command: protocol.CmdPush, Data: resp}
}

func (s *Server) handlePop(data []byte) *protocol.Frame {
	req, err := protocol.DecodePopRequest(data)
	if err != nil {
		return nil
	}

	msg := s.queueManager.Pop(req.Queue)
	if msg == nil {
		return &protocol.Frame{Command: protocol.CmdPop, Data: []byte{}}
	}

	resp := protocol.EncodePopResponse(&protocol.PopResponse{
		ID:      msg.ID,
		Payload: msg.Payload,
	})
	return &protocol.Frame{Command: protocol.CmdPop, Data: resp}
}

func (s *Server) handleAck(data []byte) *protocol.Frame {
	req, err := protocol.DecodeAckRequest(data)
	if err != nil {
		return nil
	}
	s.dedup.IsDuplicate(req.ID)
	s.queueManager.SetLastAck(req.ID)
	return &protocol.Frame{Command: protocol.CmdAck, Data: []byte{1}}
}

func (s *Server) handleStatus(data []byte) *protocol.Frame {
	req, err := protocol.DecodePopRequest(data)
	if err != nil {
		return nil
	}
	length := uint64(s.queueManager.Len(req.Queue))
	resp := protocol.EncodeStatusResponse(&protocol.StatusResponse{QueueLen: length})
	return &protocol.Frame{Command: protocol.CmdStatus, Data: resp}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) Start() error {
	go s.tcpServer.Start(s.ctx)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	go s.snapshotLoop()
	go s.delayedLoop()

	log.Printf("Server started: TCP :%d, HTTP :%d", s.config.TCPPort, s.config.HTTPPort)
	return nil
}

func (s *Server) snapshotLoop() {
	ticker := time.NewTicker(s.config.SnapshotInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			messages := s.queueManager.GetAllMessages(s.queueManager.GetLastAck())
			if err := s.snapshotter.Save(messages); err != nil {
				log.Printf("Snapshot error: %v", err)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Server) delayedLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.queueManager.MoveDelayed()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Server) Stop() error {
	s.cancel()
	s.tcpServer.Stop()
	s.pool.Shutdown()
	s.httpServer.Shutdown(context.Background())
	s.wg.Wait()
	log.Println("Server stopped")
	return nil
}
