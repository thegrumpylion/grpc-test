package pkg

import (
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)


func ParseDescriptor(path string, cc grpc.ClientConnInterface) error {

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
  
  reg.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
    for i := 0; i < fd.Services().Len(); i++ {
      svc := fd.Services().Get(i)
      fmt.Println(svc.FullName())
    }
  })

  for _, f := range fds.File {
    fmt.Println(*f.Name)
    for _, s := range f.Service {
      for _, m := range s.Method {
        fmt.Printf("/%s.%s/%s\n", *f.Package, *s.Name, *m.Name)
        fmt.Println(*m.InputType, *m.OutputType)
      }
    }
  }
  return nil
}
