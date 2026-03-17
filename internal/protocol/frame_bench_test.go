package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkEncodeFrame(b *testing.B) {
	frame := &Frame{Command: CmdPush, Data: []byte("test payload")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeFrame(frame)
	}
}

func BenchmarkDecodeFrame(b *testing.B) {
	frame := &Frame{Command: CmdPush, Data: []byte("test payload")}
	encoded := EncodeFrame(frame)
	reader := bytes.NewReader(encoded)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(encoded)
		DecodeFrame(reader)
	}
}

func BenchmarkEncodePushRequest(b *testing.B) {
	req := &PushRequest{
		Queue:    "test-queue",
		Payload:  []byte("hello world"),
		Priority: 10,
		Delay:    1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodePushRequest(req)
	}
}

func BenchmarkDecodePushRequest(b *testing.B) {
	req := &PushRequest{
		Queue:    "test-queue",
		Payload:  []byte("hello world"),
		Priority: 10,
		Delay:    1000,
	}
	encoded := EncodePushRequest(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodePushRequest(encoded)
	}
}
