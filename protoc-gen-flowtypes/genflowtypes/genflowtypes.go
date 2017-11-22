package genflowtypes

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	pbdescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/mwitkow/go-proto-validators"
)

// Dependencies maps packages names to imported type names.
type Dependencies map[string]map[string]bool

// FlowTyper is a flow language type
type FlowTyper interface {
	FlowType() string
	IsRequired() bool
}

// NamedFlowTyper is a FlowTyper with a name.
type NamedFlowTyper interface {
	FlowTyper
	Name() string
}

// knownTypeMap is a map of paths for known proto types to their desired
// flowtypes.
var knownTypeMap = map[string]string{
	".google.protobuf.Timestamp": "string",
}

func newSimpleType(typeString string, required bool) *primitiveType {
	return &primitiveType{typeString, required}
}

type primitiveType struct {
	typeString string
	required   bool
}

func (s primitiveType) FlowType() string { return s.typeString }
func (s primitiveType) IsRequired() bool { return s.required }

type messageType struct {
	*primitiveType
}

func newMessageFlowType(typeName string, required bool) *messageType {
	s := newSimpleType(typeName, required)
	return &messageType{s}
}

type repeatedType struct {
	FlowTyper
	required bool
}

func newRepeatedFlowType(underlying FlowTyper, required bool) *repeatedType {
	return &repeatedType{FlowTyper: underlying, required: required}
}

func (r repeatedType) FlowType() string { return fmt.Sprintf("Array<%s>", r.FlowTyper.FlowType()) }

type namedType struct {
	FlowTyper
	name     string
	required bool
}

func (t *namedType) Name() string {
	return t.name
}

type objectFlowType struct {
	Fields   []NamedFlowTyper
	Options  Options
	required bool
}

func (t *objectFlowType) FlowType() string {
	fields := []string{}
	for _, f := range t.Fields {
		optionalIndicator := "?"
		if f.IsRequired() {
			optionalIndicator = ""
		}
		fields = append(fields, fmt.Sprintf("  %s%s: %s", f.Name(), optionalIndicator, f.FlowType()))
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(fields, ",\n"))
}

func (t *objectFlowType) IsRequired() bool { return t.required }

func (cfg Options) fieldToType(pkg string, f *descriptor.Field, reg *descriptor.Registry, required bool) (NamedFlowTyper, Dependencies, error) {
	// FieldMessage
	var fieldType FlowTyper = newSimpleType("any", required)
	// deps will hold the other package name if the field is an external reference.
	deps := Dependencies{}
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
		fieldType = newSimpleType("number", required)
	case pbdescriptor.FieldDescriptorProto_TYPE_BOOL:
		fieldType = newSimpleType("boolean", required)
	case pbdescriptor.FieldDescriptorProto_TYPE_STRING:
		fieldType = newSimpleType("string", required)
	case pbdescriptor.FieldDescriptorProto_TYPE_GROUP:
		fieldType = newSimpleType("any", required) // , required?
	case pbdescriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: should resolve type name relative to this type
		ft, err := reg.LookupMsg("", f.GetTypeName())
		if err != nil {
			return nil, nil, err
		}
		if flowType, present := knownTypeMap[ft.FQMN()]; present {
			fieldType = newSimpleType(flowType, required)
		} else {
			fieldType = newMessageFlowType(cfg.messageTypeName(ft), required)
			if ft.File.GetPackage() != pkg {
				parts := strings.Split(ft.FQMN(), ".")
				pn := strings.Join(parts[:len(parts)-1], "")
				if _, ok := deps[pn]; !ok {
					deps[pn] = make(map[string]bool)
				}
				deps[pn][ft.GetName()] = true
			}

		}
	case pbdescriptor.FieldDescriptorProto_TYPE_BYTES:
		fieldType = newSimpleType("string", required) // could be more correct
	case pbdescriptor.FieldDescriptorProto_TYPE_ENUM:
		e, err := reg.LookupEnum("", f.GetTypeName())
		if err != nil {
			return nil, nil, err
		}

		if cfg.EmbedEnums {
			fieldType, deps, err = cfg.enumToFlowType(e, reg)
			if err != nil {
				return nil, nil, err
			}
		} else {
			name := cfg.enumTypeName(e)
			fieldType = newSimpleType(name, required)
		}
	}
	if f.GetLabel() == pbdescriptor.FieldDescriptorProto_LABEL_REPEATED {
		fieldType = newRepeatedFlowType(fieldType, required)
	}
	return &namedType{FlowTyper: fieldType, name: f.GetName(), required: required}, deps, nil
}

// this is a hack to use the FieldOptions descriptor from golang/proto instead of gogo/proto
var E_Field = &proto.ExtensionDesc{
	ExtendedType:  (*pbdescriptor.FieldOptions)(nil),
	ExtensionType: validator.E_Field.ExtensionType,
	Field:         validator.E_Field.Field,
	Name:          validator.E_Field.Name,
	Tag:           validator.E_Field.Tag,
	Filename:      validator.E_Field.Filename,
}

func init() {
	proto.RegisterType((*validator.FieldValidator)(nil), "validator.FieldValidator")
	proto.RegisterExtension(E_Field)
}

func getFieldValidatorIfAny(field *pbdescriptor.FieldDescriptorProto) *validator.FieldValidator {
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, E_Field)
		if err == nil && v.(*validator.FieldValidator) != nil {
			return (v.(*validator.FieldValidator))
		}
	}
	return nil
}

func (cfg Options) messageToFlowType(m *descriptor.Message, reg *descriptor.Registry) (FlowTyper, Dependencies, error) {
	deps := Dependencies{}
	t := &objectFlowType{
		Fields:  []NamedFlowTyper{},
		Options: cfg,
	}
	for _, f := range m.Fields {
		required := false
		if validatorOptions := getFieldValidatorIfAny(f.FieldDescriptorProto); validatorOptions != nil {
			required = *validatorOptions.MsgExists
		}
		field, newDeps, err := cfg.fieldToType(m.File.GetPackage(), f, reg, required)
		if err != nil {
			return nil, nil, err
		}
		mergeDeps(deps, newDeps)
		t.Fields = append(t.Fields, field)
	}
	return &namedType{FlowTyper: t, name: cfg.messageTypeName(m)}, deps, nil
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

func (cfg Options) enumToFlowType(e *descriptor.Enum, reg *descriptor.Registry) (FlowTyper, Dependencies, error) {
	options := []string{}
	for _, v := range e.Value {
		if !cfg.EmitEnumZeros && v.GetNumber() == 0 {
			continue
		}
		options = append(options, fmt.Sprintf(`"%s"`, v.GetName()))
	}
	name := cfg.enumTypeName(e)
	return &namedType{
		FlowTyper: newSimpleType(strings.Join(options, " | "), false),
		name:      name,
	}, nil, nil
}

func mergeDeps(dst, src Dependencies) {
	if src != nil {
		for p, types := range src {
			if _, ok := dst[p]; !ok {
				dst[p] = make(map[string]bool)
			}
			for typ, _ := range types {
				dst[p][typ] = true
			}
		}
	}
}

func generateFlowTypes(file *descriptor.File, registry *descriptor.Registry, opts Options) (string, error) {
	if opts.DumpJSON {
		m := &jsonpb.Marshaler{EmitDefaults: true, OrigName: true, Indent: "  "}
		m.Marshal(os.Stderr, file)
		time.Sleep(time.Second)
	}
	deps := Dependencies{}
	result := []FlowTyper{}
	f, err := registry.LookupFile(file.GetName())
	if err != nil {
		return "", err
	}
	for _, enum := range f.Enums {
		t, newDeps, err := opts.enumToFlowType(enum, registry)
		if err != nil {
			return "", err
		}
		mergeDeps(deps, newDeps)
		result = append(result, t)
	}
	for _, message := range f.Messages {
		t, newDeps, err := opts.messageToFlowType(message, registry)
		if err != nil {
			return "", err
		}
		mergeDeps(deps, newDeps)
		result = append(result, t)
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("").Parse(`/* @flow */
/* eslint-disable */
// Code generated by protoc-gen-flowtypes DO NOT EDIT.
// InputID: {{.InputID}}

{{range $package, $types := .Dependencies}}
import type {
{{range $type, $true := $types}}  {{$type}},
{{end -}}
} from './{{ $package }}.js';
{{end}}

{{range .Result}}export type {{.Name}} = {{.FlowType}};

{{end}}
`)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, struct {
		Options
		Dependencies Dependencies
		Result       []FlowTyper
	}{Options: opts, Dependencies: deps, Result: result})
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
