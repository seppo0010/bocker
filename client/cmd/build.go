package cmd

import (
	"context"
	"io"
	"os"
	"path"

	pb "github.com/seppo0010/bocker/protocol"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "",
	Run:   SendBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func SendBuild(cmd *cobra.Command, args []string) {
	c := cmd.Context().Value(GRPC).(pb.BockerClient)
	cwd := path.Join(cmd.Context().Value(CWD).(string), "example")
	stream, err := c.Build(context.Background(), &pb.BuildRequest{
		CwdPath:  cwd,
		FilePath: path.Join(cwd, "Bockerfile"),
		Tag:      "test",
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
