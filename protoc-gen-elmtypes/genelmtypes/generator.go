package genelmtypes

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

type generator struct {
	reg *descriptor.Registry
}

// New returns a new generator which generates elm type definition files.
func New(reg *descriptor.Registry) *generator {
	return &generator{reg: reg}
}

func (g *generator) Generate(targets []*descriptor.File, qualifyTypes bool) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File
	for _, file := range targets {
		glog.V(1).Infof("Processing %s", file.GetName())
		code, err := generateElmTypes(file, g.reg, qualifyTypes)
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
		output := fmt.Sprintf("%s.elm", strings.Title(base))
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(output),
			Content: proto.String(code),
		})
		glog.V(1).Infof("Will emit %s", output)
	}
	return files, nil
}
