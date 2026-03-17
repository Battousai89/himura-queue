package protocol

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeFrame(t *testing.T) {
	frame := &Frame{
		Command: CmdPush,
		Data:    []byte("test data"),
	}

	encoded := EncodeFrame(frame)
	decoded, err := DecodeFrame(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("DecodeFrame error: %v", err)
	}

	if decoded.Command != frame.Command {
		t.Errorf("Command mismatch: got %d, want %d", decoded.Command, frame.Command)
	}

	if !bytes.Equal(decoded.Data, frame.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, frame.Data)
	}
}

func TestEncodeDecodePushRequest(t *testing.T) {
	req := &PushRequest{
		Queue:    "test-queue",
		Payload:  []byte("hello world"),
		Priority: 10,
		Delay:    1000000000,
	}

	encoded := EncodePushRequest(req)
	decoded, err := DecodePushRequest(encoded)
	if err != nil {
		t.Fatalf("DecodePushRequest error: %v", err)
	}

	if decoded.Queue != req.Queue {
		t.Errorf("Queue mismatch: got %s, want %s", decoded.Queue, req.Queue)
	}
	if !bytes.Equal(decoded.Payload, req.Payload) {
		t.Errorf("Payload mismatch: got %v, want %v", decoded.Payload, req.Payload)
	}
	if decoded.Priority != req.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", decoded.Priority, req.Priority)
	}
	if decoded.Delay != req.Delay {
		t.Errorf("Delay mismatch: got %d, want %d", decoded.Delay, req.Delay)
	}
}

func TestEncodeDecodePopRequest(t *testing.T) {
	req := &PopRequest{Queue: "test-queue"}

	encoded := EncodePopRequest(req)
	decoded, err := DecodePopRequest(encoded)
	if err != nil {
		t.Fatalf("DecodePopRequest error: %v", err)
	}

	if decoded.Queue != req.Queue {
		t.Errorf("Queue mismatch: got %s, want %s", decoded.Queue, req.Queue)
	}
}

func TestEncodeDecodeAckRequest(t *testing.T) {
	req := &AckRequest{ID: 12345}

	encoded := EncodeAckRequest(req)
	decoded, err := DecodeAckRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeAckRequest error: %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, req.ID)
	}
}
