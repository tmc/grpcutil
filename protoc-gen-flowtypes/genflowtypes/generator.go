package genflowtypes

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

// Options describes output parameters
type Options struct {
	AlwaysQualifyTypes bool
	EmbedEnums         bool
	OptonalSimpleTypes bool
	FilenameOverride   string
	EmitEnumZeros      bool
}

// Generate processes the given proto files and produces flowtype output.
func (g *Generator) Generate(targets []*descriptor.File, opts Options) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File
	for _, file := range targets {
		glog.V(1).Infof("Processing %s", file.GetName())
		code, err := generateFlowTypes(file, g.reg, opts)
		if err == errNoTargetService {
			glog.V(1).Infof("%s: %v", file.GetName(), err)
			continue
		}
		if err != nil {
			return nil, err
		}

		name := file.GetName()
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		outputName := fmt.Sprintf("%s_types.js", base)
		if opts.FilenameOverride != "" {
			outputName = opts.FilenameOverride
		}
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(outputName),
			Content: proto.String(code),
		})
		glog.V(1).Infof("Will emit %s", outputName)
	}
	return files, nil
}
