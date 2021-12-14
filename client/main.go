package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/seppo0010/bocker/protocol"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("/tmp/bocker.sock",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBuilderClient(conn)

	stream, err := c.Build(context.Background(), &pb.BuildRequest{})
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		fmt.Printf("%#v\n", msg)
	}
}
