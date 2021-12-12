package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"sync"

	pb "github.com/seppo0010/bocker/protocol"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedBuilderServer
}

func bufferNotifier(wg *sync.WaitGroup, notify func(b []byte)) (io.ReadCloser, io.Writer) {
	read, write := io.Pipe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			b := make([]byte, 1024)
			size, err := io.ReadAtLeast(read, b, 1)
			if err != nil {
				break
			}
			notify(b[:size])
		}
	}()
	return read, write
}

func (s *server) Build(in *pb.BuildRequest, bs pb.Builder_BuildServer) error {
	cmd := exec.Command("./example.sh")
	var wg sync.WaitGroup
	var readStdout, readStderr io.ReadCloser
	readStdout, cmd.Stdout = bufferNotifier(&wg, func(b []byte) {
		bs.SendMsg(&pb.BuildReply{
			Stdout:   b,
			ExitCode: ^uint32(0),
		})
	})
	readStderr, cmd.Stderr = bufferNotifier(&wg, func(b []byte) {
		bs.SendMsg(&pb.BuildReply{
			Stderr:   b,
			ExitCode: ^uint32(0),
		})
	})
	cmd.Start()
	_ = cmd.Wait()
	readStdout.Close()
	readStderr.Close()
	wg.Wait()
	bs.SendMsg(&pb.BuildReply{ExitCode: uint32(cmd.ProcessState.ExitCode())})
	return nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 12346))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBuilderServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
