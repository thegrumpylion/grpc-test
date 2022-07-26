package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/thegrumpylion/grpc-test/pkg"
)

var testYaml = `
service: calculator.Calc
method: Add
in:
  numbers:
  - 1
  - 2
  - 3
  - 4
out:
  number: 10
`

var rootCmd = &cobra.Command{
  Use: "grpc-test",
  Args: cobra.ExactArgs(1),
  RunE: func(cmd *cobra.Command, args []string) error {

    pkg.ParseDescriptor(args[0], nil)

    return nil
  },
}

func main() {
  if err := rootCmd.Execute(); err != nil {
    log.Fatal(err)
  }
}
