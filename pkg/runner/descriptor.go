package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	Name    string
	Service string
	Method  string
	In      []json.RawMessage
	Out     []json.RawMessage
}

type testSuit struct {
	Cases []testCase
}

func TestServices(path string, cc grpc.ClientConnInterface, tests []byte) error {
	// create registry from FileDescriptorSet file

	data, err := os.ReadFile(path)
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

	ts := &testSuit{}
	if err := json.Unmarshal(testsJson, ts); err != nil {
		return err
	}

	for i, c := range ts.Cases {

		fmt.Printf("case %d: %s\n", i+1, c.Name)

		desc, err := reg.FindDescriptorByName(protoreflect.FullName(c.Service))
		if err != nil {
			return err
		}

		svc := desc.ParentFile().Services().ByName(desc.Name())
		if svc == nil {
			return fmt.Errorf("service %s not found in file %s", desc.Name(), desc.ParentFile().FullName())
		}

		meth := svc.Methods().ByName(protoreflect.Name(c.Method))
		if meth == nil {
			return fmt.Errorf("method %s not found in service %s", c.Method, svc.FullName())
		}

		in := dynamicpb.NewMessage(meth.Input())
		if err := protojson.Unmarshal(c.In[0], in); err != nil {
			return err
		}

		out := dynamicpb.NewMessage(meth.Output())
		outExpected := dynamicpb.NewMessage(meth.Output())
		if err := protojson.Unmarshal(c.Out[0], outExpected); err != nil {
			return err
		}

		if err := cc.Invoke(context.Background(), fmt.Sprintf("/%s/%s", c.Service, c.Method), in, out); err != nil {
			return err
		}

		if err := deepEqual(outExpected, out); err != nil {
			return err
		}
	}
	return nil
}

func deepEqual(expected, value *dynamicpb.Message) error {
	var err error

	expected.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !value.Has(fd) {
			err = fmt.Errorf("field not set: %s", fd.JSONName())
			return false
		}
		if !reflect.DeepEqual(v, value.Get(fd)) {
			err = fmt.Errorf("value missmatch for field %s: %s %s", fd.JSONName(), v.String(), value.Get(fd).String())
			return false
		}
		return true
	})
	return err
}
