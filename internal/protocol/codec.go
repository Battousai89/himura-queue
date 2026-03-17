package protocol

import (
	"encoding/binary"
	"errors"
)

type PushRequest struct {
	Queue    string
	Payload  []byte
	Priority int
	Delay    int64
}

type PushResponse struct {
	ID uint64
}

type PopRequest struct {
	Queue string
}

type PopResponse struct {
	ID      uint64
	Payload []byte
}

type AckRequest struct {
	ID uint64
}

type StatusResponse struct {
	QueueLen uint64
}

var (
	ErrInvalidRequest = errors.New("invalid request")
)

func EncodePushRequest(req *PushRequest) []byte {
	queueLen := uint16(len(req.Queue))
	payloadLen := uint32(len(req.Payload))
	buf := make([]byte, 2+uint32(queueLen)+4+payloadLen+4+8)
	
	pos := 0
	binary.BigEndian.PutUint16(buf[pos:], queueLen)
	pos += 2
	
	copy(buf[pos:], req.Queue)
	pos += int(queueLen)
	
	binary.BigEndian.PutUint32(buf[pos:], payloadLen)
	pos += 4
	
	copy(buf[pos:], req.Payload)
	pos += int(payloadLen)
	
	binary.BigEndian.PutUint32(buf[pos:], uint32(req.Priority))
	pos += 4
	
	binary.BigEndian.PutUint64(buf[pos:], uint64(req.Delay))
	
	return buf
}

func DecodePushRequest(data []byte) (*PushRequest, error) {
	if len(data) < 6 {
		return nil, ErrInvalidRequest
	}
	
	pos := 0
	queueLen := binary.BigEndian.Uint16(data[pos:])
	pos += 2
	
	if len(data) < pos+int(queueLen)+4 {
		return nil, ErrInvalidRequest
	}
	
	queue := string(data[pos : pos+int(queueLen)])
	pos += int(queueLen)
	
	payloadLen := binary.BigEndian.Uint32(data[pos:])
	pos += 4
	
	if len(data) < pos+int(payloadLen)+4+8 {
		return nil, ErrInvalidRequest
	}
	
	payload := make([]byte, payloadLen)
	copy(payload, data[pos:pos+int(payloadLen)])
	pos += int(payloadLen)
	
	priority := int(binary.BigEndian.Uint32(data[pos:]))
	pos += 4
	
	delay := int64(binary.BigEndian.Uint64(data[pos:]))
	
	return &PushRequest{
		Queue:    queue,
		Payload:  payload,
		Priority: priority,
		Delay:    delay,
	}, nil
}

func EncodePushResponse(resp *PushResponse) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, resp.ID)
	return buf
}

func DecodePushResponse(data []byte) (*PushResponse, error) {
	if len(data) < 8 {
		return nil, ErrInvalidRequest
	}
	return &PushResponse{
		ID: binary.BigEndian.Uint64(data),
	}, nil
}

func EncodePopRequest(req *PopRequest) []byte {
	buf := make([]byte, 2+len(req.Queue))
	binary.BigEndian.PutUint16(buf, uint16(len(req.Queue)))
	copy(buf[2:], req.Queue)
	return buf
}

func DecodePopRequest(data []byte) (*PopRequest, error) {
	if len(data) < 2 {
		return nil, ErrInvalidRequest
	}
	queueLen := binary.BigEndian.Uint16(data)
	if len(data) < 2+int(queueLen) {
		return nil, ErrInvalidRequest
	}
	return &PopRequest{
		Queue: string(data[2 : 2+queueLen]),
	}, nil
}

func EncodePopResponse(resp *PopResponse) []byte {
	buf := make([]byte, 8+4+len(resp.Payload))
	binary.BigEndian.PutUint64(buf, resp.ID)
	binary.BigEndian.PutUint32(buf[8:], uint32(len(resp.Payload)))
	copy(buf[12:], resp.Payload)
	return buf
}

func DecodePopResponse(data []byte) (*PopResponse, error) {
	if len(data) < 12 {
		return nil, ErrInvalidRequest
	}
	id := binary.BigEndian.Uint64(data)
	payloadLen := binary.BigEndian.Uint32(data[8:])
	if len(data) < 12+int(payloadLen) {
		return nil, ErrInvalidRequest
	}
	payload := make([]byte, payloadLen)
	copy(payload, data[12:])
	return &PopResponse{
		ID:      id,
		Payload: payload,
	}, nil
}

func EncodeAckRequest(req *AckRequest) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, req.ID)
	return buf
}

func DecodeAckRequest(data []byte) (*AckRequest, error) {
	if len(data) < 8 {
		return nil, ErrInvalidRequest
	}
	return &AckRequest{
		ID: binary.BigEndian.Uint64(data),
	}, nil
}

func EncodeStatusResponse(resp *StatusResponse) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, resp.QueueLen)
	return buf
}

func DecodeStatusResponse(data []byte) (*StatusResponse, error) {
	if len(data) < 8 {
		return nil, ErrInvalidRequest
	}
	return &StatusResponse{
		QueueLen: binary.BigEndian.Uint64(data),
	}, nil
}
