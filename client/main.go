package main

import (
	"context"
	"net"
	"os"

	"github.com/seppo0010/bocker/client/cmd"
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
	c := pb.NewBockerClient(conn)

	cwd, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to get cwd")
	}

	ctx := context.WithValue(
		context.WithValue(context.Background(), cmd.GRPC, c), cmd.CWD, cwd)
	cmd.ExecuteContext(ctx)
}
