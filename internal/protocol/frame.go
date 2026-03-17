package protocol

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	HeaderSize = 5
)

type CommandType byte

const (
	CmdPush      CommandType = 1
	CmdPop       CommandType = 2
	CmdAck       CommandType = 3
	CmdSubscribe CommandType = 4
	CmdStatus    CommandType = 5
)

var (
	ErrInvalidFrame = errors.New("invalid frame")
	ErrEOF          = errors.New("unexpected EOF")
)

type Frame struct {
	Command CommandType
	Data    []byte
}

func EncodeFrame(frame *Frame) []byte {
	buf := make([]byte, HeaderSize+len(frame.Data))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(frame.Data)))
	buf[4] = byte(frame.Command)
	copy(buf[HeaderSize:], frame.Data)
	return buf
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrEOF
		}
		return nil, err
	}

	dataLen := binary.BigEndian.Uint32(header[0:4])
	cmd := CommandType(header[4])

	data := make([]byte, dataLen)
	if dataLen > 0 {
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
	}

	return &Frame{Command: cmd, Data: data}, nil
}
