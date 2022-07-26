package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"sigs.k8s.io/yaml"
)

type testCase struct {
  Service string
  Method string
  In json.RawMessage
  Out json.RawMessage
}

func ParseDescriptor(path string, cc grpc.ClientConnInterface, tests []byte) error {

  data, err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }

  testsJson, err := yaml.YAMLToJSON(tests)
  if err != nil {
    return err
  }

  tc := &testCase{}
  if err := json.Unmarshal(testsJson, tc); err != nil {
    return err
  }

  fds := &descriptorpb.FileDescriptorSet{}
  if err := proto.Unmarshal(data, fds); err != nil {
    return err
  }

  reg, err := protodesc.NewFiles(fds)
  if err != nil {
    return err
  }
  
  reg.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
    for i := 0; i < fd.Services().Len(); i++ {
      svc := fd.Services().Get(i)
      fmt.Println(svc.FullName())
      for j := 0; j < svc.Methods().Len(); j++ {
        mth := svc.Methods().Get(j)
        fmt.Println(mth.FullName(), mth.Name())
      }
    }
    for i := 0; i < fd.Messages().Len(); i++ {
      msg := fd.Messages().Get(i)
      fmt.Println(msg.FullName())
    }
    return true
  })

  return nil
}
