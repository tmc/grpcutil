package genflowtypes

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

var (
	errNoTargetService = errors.New("no target service defined in the file")
)

type generator struct {
	reg *descriptor.Registry
}

// New returns a new generator which generates flowtype type definition files.
func New(reg *descriptor.Registry) *generator {
	return &generator{reg: reg}
}

func (g *generator) Generate(targets []*descriptor.File, qualifyTypes bool, embedEnums bool) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File
	for _, file := range targets {
		glog.V(1).Infof("Processing %s", file.GetName())
		code, err := generateFlowTypes(file, g.reg, qualifyTypes, embedEnums)
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
		output := fmt.Sprintf("%sTypes.js", base)
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(output),
			Content: proto.String(code),
		})
		glog.V(1).Infof("Will emit %s", output)
	}
	return files, nil
}
