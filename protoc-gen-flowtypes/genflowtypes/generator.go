package genflowtypes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/pkg/errors"
	"github.com/tmc/grpcutil/protoc-gen-flowtypes/opts"
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
	AlwaysQualifyTypes bool
	EmbedEnums         bool
	OptonalSimpleTypes bool
	FilenameOverride   string
	EmitEnumZeros      bool
	InputID            string
	DumpJSON           bool
	ProtoOptions       opts.Options
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
	if opts.FilenameOverride != "" {
		outputNames = strings.Split(opts.FilenameOverride, "+")
	}
	for i, file := range targets {
		glog.V(1).Infof("Processing %s", file.GetName())
		code, err := generateFlowTypes(file, g.reg, opts)
		if err == errNoTargetService {
			glog.V(1).Infof("%s: %v", file.GetName(), err)
			continue
		}
		if err != nil {
			return nil, errors.Wrap(err, "generateFlowTypes")
		}

		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(outputNames[i]),
			Content: proto.String(code),
		})
		glog.V(1).Infof("Will emit %s", outputNames[i])
	}
	return files, nil
}
