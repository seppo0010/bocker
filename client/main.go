package main

import (
	"context"
	"io"
	"net"
	"os"

	pb "github.com/seppo0010/bocker/protocol"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("/tmp/bocker.sock",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to connect")
	}
	defer conn.Close()
	c := pb.NewBuilderClient(conn)

	stream, err := c.Build(context.Background(), &pb.BuildRequest{})
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		os.Stderr.Write(msg.Stderr)
		os.Stdout.Write(msg.Stdout)
		if msg.ExitCode != ^uint32(0) {
			os.Exit(int(msg.ExitCode))
		}
	}
}
