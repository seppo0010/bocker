package main

import (
	"net"
	"os"
	"path"

	bocker "github.com/seppo0010/bocker/protocol"
	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/server/build"
	"github.com/seppo0010/bocker/server/run"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedBockerServer
}

func getConfig() (*shared.Config, func(), error) {
	bockerBuildPath, err := os.MkdirTemp("", "bocker-build-path")
	if err != nil {
		return nil, nil, err
	}

	return &shared.Config{
			BockerPath: path.Join(os.Getenv("HOME"), ".bocker"),
			BuildPath:  bockerBuildPath,
		}, func() {
			os.RemoveAll(bockerBuildPath)
		}, nil
}

func (s *Server) createReplyChannel(bs bocker.Bocker_RunServer) chan<- *pb.ExecReply {
	sendMessageChan := make(chan *pb.ExecReply)
	go func() {
		for message := range sendMessageChan {
			err := bs.SendMsg(message)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to send message")
			}
		}
	}()
	return sendMessageChan
}

func (s *Server) Build(in *pb.BuildRequest, bs bocker.Bocker_BuildServer) error {
	conf, cleanup, err := getConfig()
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = build.Run(in, conf)
	return err
}

func (s *Server) Run(in *pb.RunRequest, bs bocker.Bocker_RunServer) error {
	conf, cleanup, err := getConfig()
	if err != nil {
		return err
	}
	defer cleanup()
	outchan := s.createReplyChannel(bs)
	err = run.Run(in, conf, outchan)
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
	pb.RegisterBockerServer(s, &Server{})
	log.WithFields(log.Fields{
		"address": lis.Addr(),
	}).Info("server listening")
	if err := s.Serve(lis); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to serve")
	}
}
