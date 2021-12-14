package main

import (
	"io"
	"os/exec"
	"sync"

	log "github.com/sirupsen/logrus"
)

func bufferNotifier(wg *sync.WaitGroup, notify func(b []byte) error) (io.ReadCloser, io.Writer) {
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
			err = notify(b[:size])
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to notify")
				break
			}
		}
	}()
	return read, write
}

func (s *Server) Build(in *pb.BuildRequest, bs pb.Builder_BuildServer) error {
	cmd := exec.Command("./example.sh")
	var wg sync.WaitGroup
	var readStdout, readStderr io.ReadCloser
	sendMessageChan := make(chan *pb.BuildReply)
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
	readStdout, cmd.Stdout = bufferNotifier(&wg, func(b []byte) error {
		sendMessageChan <- &pb.BuildReply{
			Stdout:   b,
			ExitCode: ^uint32(0),
		}
		return nil
	})
	readStderr, cmd.Stderr = bufferNotifier(&wg, func(b []byte) error {
		sendMessageChan <- &pb.BuildReply{
			Stderr:   b,
			ExitCode: ^uint32(0),
		}
		return nil
	})
	cmd.Start()
	_ = cmd.Wait()
	readStdout.Close()
	readStderr.Close()
	wg.Wait()
	close(sendMessageChan)
	bs.SendMsg(&pb.BuildReply{ExitCode: uint32(cmd.ProcessState.ExitCode())})
	return nil
}
