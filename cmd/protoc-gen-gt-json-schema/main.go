package main

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

var logger *slog.Logger

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Read the request from protoc (stdin)
	request, err := io.ReadAll(os.Stdin)
	if err != nil {
		logger.Error("failed to read request", "err", err)
		os.Exit(1)
	}

	// Parse the request
	var req pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(request, &req); err != nil {
		logger.Error("failed to unmarshal request", "err", err)
		os.Exit(1)
	}

	// Initialize the protogen plugin
	gen, err := protogen.Options{}.New(&req)
	if err != nil {
		logger.Error("failed to initialize protogen plugin", "err", err)
		os.Exit(1)
	}

	// Generate the JSON schema for each proto file
	var files []*pluginpb.CodeGeneratorResponse_File
	for _, file := range gen.Files {
		if file.Generate {
			files = append(files, generateJSONSchemaFile(file))
		}
	}

	// Create a response and marshal it to stdout
	response := &pluginpb.CodeGeneratorResponse{File: files}
	output, err := proto.Marshal(response)
	if err != nil {
		logger.Error("failed to marshal response", "err", err)
		os.Exit(1)
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		logger.Error("failed to write response", "err", err)
		os.Exit(1)
	}
}

func generateJSONSchemaFile(file *protogen.File) *pluginpb.CodeGeneratorResponse_File {
	logger.Info("generating JSON schema for file", "file", file.Desc.Path())

	// This function will generate JSON schema for each service and message.
	filename := file.GeneratedFilenamePrefix + ".schema.json"
	content := generateJSONSchema(file)

	return &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(filename),
		Content: proto.String(content),
	}
}

func generateJSONSchema(file *protogen.File) string {
	logger.Info("generating JSON schema", "file", file.Desc.Path())
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       file.Desc.Path(),
		"type":        "object",
		"definitions": make(map[string]interface{}),
		"properties":  make(map[string]interface{}),
	}

	// Loop through each service and message
	for _, svc := range file.Services {
		generateServiceSchema(schema, svc)
	}

	for _, msg := range file.Messages {
		generateMessageSchema(schema, msg)
	}

	// Marshal the schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	return string(schemaJSON)
}

func generateServiceSchema(schema map[string]interface{}, svc *protogen.Service) {
	logger.Info("generating JSON schema for service", "service", svc.Desc.Name())
	properties := schema["properties"].(map[string]interface{})
	serviceSchema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	for _, method := range svc.Methods {
		methodSchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"in": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$ref": "#/definitions/" + string(method.Input.Desc.Name()),
					},
				},
				"out": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"$ref": "#/definitions/" + string(method.Output.Desc.Name()),
					},
				},
			},
		}

		serviceSchema["properties"].(map[string]interface{})[string(method.Desc.Name())] = methodSchema
	}

	properties[string(svc.Desc.Name())] = serviceSchema
}

func generateMessageSchema(schema map[string]interface{}, msg *protogen.Message) {
	logger.Info("generating JSON schema for message", "message", msg.Desc.Name())
	definitions := schema["definitions"].(map[string]interface{})
	messageSchema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	for _, field := range msg.Fields {
		fieldSchema := map[string]interface{}{
			"type": fieldTypeToJSONType(field.Desc.Kind()),
		}
		messageSchema["properties"].(map[string]interface{})[string(field.Desc.Name())] = fieldSchema
	}

	definitions[string(msg.Desc.Name())] = messageSchema
}

func fieldTypeToJSONType(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return "integer"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "number"
	default:
		return "string" // Default to string for unknown types
	}
}
