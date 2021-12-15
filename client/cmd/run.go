package cmd

import (
	"context"
	"io"
	"os"

	pb "github.com/seppo0010/bocker/protocol"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Run:   SendRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func SendRun(cmd *cobra.Command, args []string) {
	c := cmd.Context().Value(GRPC).(pb.BockerClient)
	stream, err := c.Run(context.Background(), &pb.RunRequest{
		Tag: "test",
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalf("failed to send request")
	}
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
