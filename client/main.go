package main

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/seppo0010/bocker/protocol"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:12346", grpc.WithInsecure())
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
