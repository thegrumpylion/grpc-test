package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"sigs.k8s.io/yaml"
)

type testCase struct {
  Service string
  Method string
  In json.RawMessage
  Out json.RawMessage
}

func TestServices(path string, cc grpc.ClientConnInterface, tests []byte) error {
  
  // create registry from FileDescriptorSet file

  data, err := ioutil.ReadFile(path)
  if err != nil {
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

  // create test cases
  
  testsJson, err := yaml.YAMLToJSON(tests)
  if err != nil {
    return err
  }

  tc := &testCase{}
  if err := json.Unmarshal(testsJson, tc); err != nil {
    return err
  }

  desc, err := reg.FindDescriptorByName(protoreflect.FullName(tc.Service))
  if err != nil {
    return err
  }

  svc := desc.ParentFile().Services().ByName(desc.Name())
  if svc == nil {
    return fmt.Errorf("service %s not found in file %s", desc.Name(), desc.ParentFile().FullName())
  }

  meth := svc.Methods().ByName(protoreflect.Name(tc.Method))
  if meth == nil {
    return fmt.Errorf("method %s not found in service %s", tc.Method, svc.FullName())
  }

  in := dynamicpb.NewMessage(meth.Input())
  if err := protojson.Unmarshal(tc.In, in); err != nil {
    return err
  }

  out := dynamicpb.NewMessage(meth.Output())
  outExpected := dynamicpb.NewMessage(meth.Output())
  if err := protojson.Unmarshal(tc.Out, outExpected); err != nil {
    return err
  }

  if err := cc.Invoke(context.Background(),fmt.Sprintf("/%s/%s", tc.Service, tc.Method), in, out); err != nil {
    return err
  }

  outExpected.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
    if !out.Has(fd) {
      err = fmt.Errorf("field not set: %s", fd.JSONName())
      return false
    }
    if !reflect.DeepEqual(v, out.Get(fd)) {
      err = fmt.Errorf("value missmatch for field %s: %s %s", fd.JSONName(), v.String(), out.Get(fd).String())
      return false
    }
    return true
  })



  return err
}
