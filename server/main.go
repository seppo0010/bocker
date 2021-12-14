package main

import (
	"net"

	pb "github.com/seppo0010/bocker/protocol"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedBuilderServer
}

func main() {
	lis, err := net.Listen("unix", "/tmp/bocker.sock")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to listen")
	}
	s := grpc.NewServer()
	pb.RegisterBuilderServer(s, &Server{})
	log.WithFields(log.Fields{
		"address": lis.Addr(),
	}).Info("server listening")
	if err := s.Serve(lis); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to serve")
	}
}
