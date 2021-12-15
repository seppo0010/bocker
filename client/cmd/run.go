package cmd

import (
	"context"

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
	readOutput(stream)
}
