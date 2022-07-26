package main

import (
	"context"
	"log"
	"net"

	testpb "github.com/thegrumpylion/grpc-test/test/proto"
	"google.golang.org/grpc"
)

type caclService struct {
  testpb.UnimplementedCalcServer
}


func (s *caclService) Add(ctx context.Context, in *testpb.IntList) (*testpb.Int, error) {
  ret := &testpb.Int{}
  for _, i := range in.Numbers {
    ret.Number += i
  }
  return ret, nil
}

func (s *caclService) Sub(ctx context.Context, in *testpb.IntList) (*testpb.Int, error) {
  ret := &testpb.Int{}
  for _, i := range in.Numbers {
    ret.Number -= i
  }
  return ret, nil
}

func main() {

  lis, err := net.Listen("tcp", ":5051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
  
  s := grpc.NewServer()

  testpb.RegisterCalcServer(s, &caclService{})

  if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
