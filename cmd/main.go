package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/thegrumpylion/grpc-test/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

    conn, err := grpc.Dial(":5051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
      log.Fatalf("did not connect: %v", err)
    }

    return pkg.TestServices(args[0], conn, []byte(testYaml))
  },
}

func main() {
  if err := rootCmd.Execute(); err != nil {
    log.Fatal(err)
  }
}
