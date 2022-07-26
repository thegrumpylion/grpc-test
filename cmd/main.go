package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
  Use: "grpc-test",
  RunE: func(cmd *cobra.Command, args []string) error {
    return nil
  },
}

func main() {
  
}
