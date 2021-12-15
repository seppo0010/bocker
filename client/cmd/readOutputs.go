package cmd

import (
	"io"
	"os"

	pb "github.com/seppo0010/bocker/protocol"
	log "github.com/sirupsen/logrus"
)

func readOutput(stream interface {
	Recv() (*pb.ExecReply, error)
}) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalf("failed to receive message")
		}
		os.Stderr.Write(msg.Stderr)
		os.Stdout.Write(msg.Stdout)
		if msg.ExitCode != ^uint32(0) {
			os.Exit(int(msg.ExitCode))
		}
	}
}
