package persistence

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
	"time"

	"himura-queue/pkg/models"
)

type Snapshotter struct {
	path     string
	interval time.Duration
	mu       sync.Mutex
}

func NewSnapshotter(path string, interval time.Duration) *Snapshotter {
	return &Snapshotter{
		path:     path,
		interval: interval,
	}
}

func (s *Snapshotter) Save(messages []*models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tmpPath := s.path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)
	if err := s.writeSnapshot(f, messages); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmpPath, s.path)
}

func (s *Snapshotter) writeSnapshot(w io.Writer, messages []*models.Message) error {
	count := uint32(len(messages))
	if err := binary.Write(w, binary.BigEndian, count); err != nil {
		return err
	}
	for _, msg := range messages {
		if err := s.writeMessage(w, msg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Snapshotter) writeMessage(w io.Writer, msg *models.Message) error {
	if err := binary.Write(w, binary.BigEndian, msg.ID); err != nil {
		return err
	}
	queueLen := uint16(len(msg.Queue))
	if err := binary.Write(w, binary.BigEndian, queueLen); err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg.Queue)); err != nil {
		return err
	}
	payloadLen := uint32(len(msg.Payload))
	if err := binary.Write(w, binary.BigEndian, payloadLen); err != nil {
		return err
	}
	if _, err := w.Write(msg.Payload); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(msg.Priority)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, int64(msg.Delay)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, msg.CreatedAt.UnixNano()); err != nil {
		return err
	}
	return nil
}

func (s *Snapshotter) Load() ([]*models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	return s.readSnapshot(f)
}

func (s *Snapshotter) readSnapshot(r io.Reader) ([]*models.Message, error) {
	var count uint32
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return nil, err
	}
	messages := make([]*models.Message, 0, count)
	for i := uint32(0); i < count; i++ {
		msg, err := s.readMessage(r)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *Snapshotter) readMessage(r io.Reader) (*models.Message, error) {
	msg := &models.Message{}
	if err := binary.Read(r, binary.BigEndian, &msg.ID); err != nil {
		return nil, err
	}
	var queueLen uint16
	if err := binary.Read(r, binary.BigEndian, &queueLen); err != nil {
		return nil, err
	}
	queueBuf := make([]byte, queueLen)
	if _, err := r.Read(queueBuf); err != nil {
		return nil, err
	}
	msg.Queue = string(queueBuf)
	var payloadLen uint32
	if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
		return nil, err
	}
	msg.Payload = make([]byte, payloadLen)
	if _, err := r.Read(msg.Payload); err != nil {
		return nil, err
	}
	var priority uint32
	if err := binary.Read(r, binary.BigEndian, &priority); err != nil {
		return nil, err
	}
	msg.Priority = int(priority)
	var delay int64
	if err := binary.Read(r, binary.BigEndian, &delay); err != nil {
		return nil, err
	}
	msg.Delay = time.Duration(delay)
	var createdAt int64
	if err := binary.Read(r, binary.BigEndian, &createdAt); err != nil {
		return nil, err
	}
	msg.CreatedAt = time.Unix(0, createdAt)
	return msg, nil
}
