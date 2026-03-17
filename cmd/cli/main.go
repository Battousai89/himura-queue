package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"

	"himura-queue/internal/protocol"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "push":
		handlePush(os.Args[2:])
	case "pop":
		handlePop(os.Args[2:])
	case "stats":
		handleStats(os.Args[2:])
	case "health":
		handleHealth(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: himura-cli <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  push   - Push message to queue")
	fmt.Println("  pop    - Pop message from queue")
	fmt.Println("  stats  - Get queue statistics")
	fmt.Println("  health - Check server health")
}

func dial(host string, port int) (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
}

func handlePush(args []string) {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	host := fs.String("host", "localhost", "Server host")
	port := fs.Int("port", 9000, "Server port")
	queue := fs.String("queue", "default", "Queue name")
	payload := fs.String("payload", "", "Message payload")
	priority := fs.Int("priority", 0, "Message priority")
	delay := fs.Duration("delay", 0, "Message delay")
	fs.Parse(args)

	conn, err := dial(*host, *port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	req := protocol.EncodePushRequest(&protocol.PushRequest{
		Queue:    *queue,
		Payload:  []byte(*payload),
		Priority: *priority,
		Delay:    int64(*delay),
	})

	frame := &protocol.Frame{Command: protocol.CmdPush, Data: req}
	if _, err := conn.Write(protocol.EncodeFrame(frame)); err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		os.Exit(1)
	}

	resp, err := protocol.DecodeFrame(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		os.Exit(1)
	}

	pushResp, _ := protocol.DecodePushResponse(resp.Data)
	fmt.Printf("Message pushed with ID: %d\n", pushResp.ID)
}

func handlePop(args []string) {
	fs := flag.NewFlagSet("pop", flag.ExitOnError)
	host := fs.String("host", "localhost", "Server host")
	port := fs.Int("port", 9000, "Server port")
	queue := fs.String("queue", "default", "Queue name")
	fs.Parse(args)

	conn, err := dial(*host, *port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	req := protocol.EncodePopRequest(&protocol.PopRequest{Queue: *queue})
	frame := &protocol.Frame{Command: protocol.CmdPop, Data: req}
	if _, err := conn.Write(protocol.EncodeFrame(frame)); err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		os.Exit(1)
	}

	resp, err := protocol.DecodeFrame(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Data) == 0 {
		fmt.Println("No messages")
		return
	}

	popResp, _ := protocol.DecodePopResponse(resp.Data)
	fmt.Printf("Message ID: %d, Payload: %s\n", popResp.ID, string(popResp.Payload))
}

func handleStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	host := fs.String("host", "localhost", "Server host")
	port := fs.Int("port", 9000, "Server port")
	queue := fs.String("queue", "default", "Queue name")
	fs.Parse(args)

	conn, err := dial(*host, *port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	data := make([]byte, 2+len(*queue))
	binary.BigEndian.PutUint16(data, uint16(len(*queue)))
	copy(data[2:], *queue)

	frame := &protocol.Frame{Command: protocol.CmdStatus, Data: data}
	if _, err := conn.Write(protocol.EncodeFrame(frame)); err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		os.Exit(1)
	}

	resp, err := protocol.DecodeFrame(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		os.Exit(1)
	}

	statusResp, _ := protocol.DecodeStatusResponse(resp.Data)
	fmt.Printf("Queue '%s' length: %d\n", *queue, statusResp.QueueLen)
}

func handleHealth(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	host := fs.String("host", "localhost", "Server host")
	httpPort := fs.Int("http-port", 9001, "HTTP port")
	fs.Parse(args)

	conn, err := net.Dial("tcp", fmt.Sprintf("[%s]:%d", *host, *httpPort))
	if err != nil {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *httpPort))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
		os.Exit(1)
	}
	conn.Close()

	fmt.Println("Server is healthy")
}
