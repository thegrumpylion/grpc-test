package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/thegrumpylion/grpc-test/pkg/runner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var testYaml = `
cases:
-
  name: add 1 2 3 4 expect 10
  service: calculator.Calc
  method: Add
  in:
  - numbers:
    - 1
    - 2
    - 3
    - 4
  out:
  - number: 10
-
  name: add 1 2 3 4 expect 12 and break
  service: calculator.Calc
  method: Add
  in:
  - numbers:
    - 1
    - 2
    - 3
    - 4
  out:
  - number: 12
`

var rootCmd = &cobra.Command{
	Use:  "grpc-test",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		conn, err := grpc.Dial(":5051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		if err = runner.TestServices(args[0], conn, []byte(testYaml)); err != nil {
			fmt.Println(err)
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
