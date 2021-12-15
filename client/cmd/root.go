package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

const GRPC = "gRPC"
const CWD = "CWD"

var rootCmd = &cobra.Command{
	Use: "bocker",
}

func ExecuteContext(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
