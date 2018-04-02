package genjsrpc

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

var (
	errNoTargetService = errors.New("no target service defined in the file")
)

// Generator processes proto descriptors and generates flow type definitions.
type Generator struct {
	reg *descriptor.Registry
}

// New returns a new generator which generates flowtype type definition files.
func New(reg *descriptor.Registry) *Generator {
	return &Generator{reg: reg}
}

// GeneratorOptions describes output parameters
type GeneratorOptions struct {
	InputID string
}

func defaultOutputNames(targets []*descriptor.File) []string {
	result := []string{}
	for _, file := range targets {
		name := file.GetName()
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		result = append(result, fmt.Sprintf("%s.js", base))
	}
	return result
}

// Generate processes the given proto files and produces flowtype output.
func (g *Generator) Generate(targets []*descriptor.File, opts GeneratorOptions) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File
	outputNames := defaultOutputNames(targets)
	for i, file := range targets {
		glog.V(1).Infof("Processing %s", file.GetName())
		code, err := generateJSRPC(file, g.reg, opts)
		if err == errNoTargetService {
			glog.V(1).Infof("%s: %v", file.GetName(), err)
			continue
		}
		if err != nil {
			return nil, err
		}

		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(outputNames[i]),
			Content: proto.String(code),
		})
		glog.V(1).Infof("Will emit %s", outputNames[i])
	}
	return files, nil
}
