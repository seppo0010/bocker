package main

import (
	"fmt"
	"net"
	"os"

	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/server/build"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedBuilderServer
}

func getConfig() (*shared.Config, func(), error) {
	bockerTmpdir, err := os.MkdirTemp("", "bocker-test")
	if err != nil {
		return nil, nil, err
	}

	bockerBuildPath, err := os.MkdirTemp("", "bocker-build-path")
	if err != nil {
		return nil, nil, err
	}

	return &shared.Config{
			BockerPath: bockerTmpdir,
			BuildPath:  bockerBuildPath,
		}, func() {
			os.RemoveAll(bockerTmpdir)
			os.RemoveAll(bockerBuildPath)
		}, nil
}

func (s *Server) Build(in *pb.BuildRequest, bs pb.Builder_BuildServer) error {
	fmt.Printf("%#v\n", in)
	getConfig()
	conf, cleanup, err := getConfig()
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = build.Run(in, conf)
	return err
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
