package genflowtypes

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	pbdescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

type FlowType interface {
	FlowType() string
}
type NamedFlowType interface {
	FlowType
	FlowTypeName() string
}

type simpleFlowType string

func (s simpleFlowType) FlowType() string { return string(s) }

type messageFlowType string

func (s messageFlowType) FlowType() string { return string(s) }

type repeatedFlowType struct {
	t FlowType
}

func (r repeatedFlowType) FlowType() string { return fmt.Sprintf("Array<%s>", r.t.FlowType()) }

type namedFlowType struct {
	Name string
	Type FlowType
}

func (t *namedFlowType) FlowType() string {
	return t.Type.FlowType()
	return fmt.Sprintf("%s = %s", t.Name, t.Type.FlowType())
}
func (t *namedFlowType) FlowTypeName() string {
	return t.Name
}

type objectFlowType struct {
	Fields  []NamedFlowType
	Options Options
}

func (t *objectFlowType) FlowType() string {
	fields := []string{}
	for _, f := range t.Fields {
		optionalIndicator := "?"
		if _, simple := f.(*namedFlowType).Type.(simpleFlowType); simple && t.Options.OptonalSimpleTypes == false {
			optionalIndicator = ""
		}
		fields = append(fields, fmt.Sprintf("  %s%s: %s", f.FlowTypeName(), optionalIndicator, f.FlowType()))
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(fields, ",\n"))
}

func (cfg Options) fqmnToType(fqmn string, registry *descriptor.Registry) (FlowType, error) {
	m, err := registry.LookupMsg("", fqmn)
	if err != nil {
		return nil, err
	}
	return cfg.messageToFlowType(m, registry)
}

func (cfg Options) fieldToType(f *descriptor.Field, reg *descriptor.Registry) (NamedFlowType, error) {
	// FieldMessage
	var fieldType FlowType = simpleFlowType("any")
	switch f.GetType() {
	case pbdescriptor.FieldDescriptorProto_TYPE_DOUBLE:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_FLOAT:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_INT64:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_UINT64:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_INT32:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_FIXED64:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_FIXED32:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_UINT32:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_SFIXED32:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_SFIXED64:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_SINT32:
		fallthrough
	case pbdescriptor.FieldDescriptorProto_TYPE_SINT64:
		fieldType = simpleFlowType("number")
	case pbdescriptor.FieldDescriptorProto_TYPE_BOOL:
		fieldType = simpleFlowType("boolean")
	case pbdescriptor.FieldDescriptorProto_TYPE_STRING:
		fieldType = simpleFlowType("string")
	case pbdescriptor.FieldDescriptorProto_TYPE_GROUP:
		fieldType = simpleFlowType("any") // ?
	case pbdescriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: should resolve type name relative to this type
		ft, err := reg.LookupMsg("", f.GetTypeName())
		if err != nil {
			return nil, err
		}
		fieldType = messageFlowType(cfg.messageTypeName(ft))
	case pbdescriptor.FieldDescriptorProto_TYPE_BYTES:
		fieldType = simpleFlowType("string") // could be more correct
	case pbdescriptor.FieldDescriptorProto_TYPE_ENUM:
		e, err := reg.LookupEnum("", f.GetTypeName())
		if err != nil {
			return nil, err
		}

		if cfg.EmbedEnums {
			fieldType, err = cfg.enumToFlowType(e, reg)
			if err != nil {
				return nil, err
			}
		} else {
			name := cfg.enumTypeName(e)
			fieldType = simpleFlowType(name)
		}
	}
	if f.GetLabel() == pbdescriptor.FieldDescriptorProto_LABEL_REPEATED {
		fieldType = repeatedFlowType{fieldType}
	}
	return &namedFlowType{Name: f.GetName(), Type: fieldType}, nil
}

func (cfg Options) messageToFlowType(m *descriptor.Message, reg *descriptor.Registry) (FlowType, error) {
	t := &objectFlowType{
		Fields:  []NamedFlowType{},
		Options: cfg,
	}
	for _, f := range m.Fields {
		field, err := cfg.fieldToType(f, reg)
		if err != nil {
			return nil, err
		}
		t.Fields = append(t.Fields, field)
	}
	return &namedFlowType{Name: cfg.messageTypeName(m), Type: t}, nil
}

func (cfg Options) enumTypeName(e *descriptor.Enum) string {
	name := strings.Replace(e.FQEN(), ".", "", -1)
	if !cfg.AlwaysQualifyTypes {
		if strings.HasPrefix(name, e.File.GoPkg.Name) {
			name = name[len(e.File.GoPkg.Name):]
		}
	}
	return name
}

func (cfg Options) messageTypeName(m *descriptor.Message) string {
	name := strings.Replace(m.FQMN(), ".", "", -1)
	if !cfg.AlwaysQualifyTypes {
		if strings.HasPrefix(name, m.File.GoPkg.Name) {
			name = name[len(m.File.GoPkg.Name):]
		}
	}
	return name
}

func (cfg Options) enumToFlowType(e *descriptor.Enum, reg *descriptor.Registry) (FlowType, error) {
	options := []string{}
	for _, v := range e.Value {
		options = append(options, fmt.Sprintf(`"%s"`, v.GetName()))
	}
	name := cfg.enumTypeName(e)
	return &namedFlowType{
		Name: name,
		Type: simpleFlowType(strings.Join(options, " | ")),
	}, nil
}

func generateFlowTypes(file *descriptor.File, registry *descriptor.Registry, opts Options) (string, error) {
	result := []FlowType{}
	f, err := registry.LookupFile(file.GetName())
	if err != nil {
		return "", err
	}
	for _, enum := range f.Enums {
		t, err := opts.enumToFlowType(enum, registry)
		if err != nil {
			return "", err
		}
		result = append(result, t)
	}
	for _, message := range f.Messages {
		t, err := opts.messageToFlowType(message, registry)
		if err != nil {
			return "", err
		}
		result = append(result, t)
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("").Parse("/* @flow */\n// Code generated by protoc-gen-flowtypes DO NOT EDIT.\n\n{{range .}}export type {{.FlowTypeName}} = {{.FlowType}};\n\n{{end}}")
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, result)
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
