package genflowtypes

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	pbdescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/pkg/errors"
	"github.com/tmc/grpcutil/protoc-gen-flowtypes/opts"
)

// Dependencies maps packages names to imported type names.
type Dependencies map[string]map[string]bool

// FlowTyper is a flow language type
type FlowTyper interface {
	FlowType() string
	IsRequired() bool
	IsNullable() bool
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

func newSimpleType(typeString string, opts opts.Options) *primitiveType {
	return &primitiveType{typeString, opts}
}

type primitiveType struct {
	typeString string
	opts       opts.Options
}

func (s primitiveType) FlowType() string { return s.typeString }
func (s primitiveType) IsRequired() bool { return s.opts.GetRequired() }
func (s primitiveType) IsNullable() bool { return s.opts.GetNullable() }

type messageType struct {
	*primitiveType
}

func newMessageFlowType(typeName string, opts opts.Options) *messageType {
	s := newSimpleType(typeName, opts)
	return &messageType{s}
}

type repeatedType struct {
	FlowTyper
	opts opts.Options
}

func newRepeatedFlowType(underlying FlowTyper, opts opts.Options) *repeatedType {
	return &repeatedType{FlowTyper: underlying, opts: opts}
}

func (r repeatedType) FlowType() string { return fmt.Sprintf("Array<%s>", r.FlowTyper.FlowType()) }

type namedType struct {
	FlowTyper
	name string
	opts opts.Options
}

func (t *namedType) Name() string {
	return t.name
}

type objectFlowType struct {
	Fields  []NamedFlowTyper
	Options GeneratorOptions

	opts opts.Options
}

func (t *objectFlowType) FlowType() string {
	fields := []string{}
	for _, f := range t.Fields {
		optionalIndicator := "?"
		nullableIndicator := "?"
		if f.IsRequired() {
			optionalIndicator = ""
		}
		if !f.IsNullable() {
			nullableIndicator = ""
		}
		fields = append(fields, fmt.Sprintf("  %s%s: %s%s", f.Name(), optionalIndicator, nullableIndicator, f.FlowType()))
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(fields, ",\n"))
}

func (t *objectFlowType) IsRequired() bool { return t.opts.GetRequired() }
func (t *objectFlowType) IsNullable() bool { return t.opts.GetNullable() }

func (cfg GeneratorOptions) fieldToType(pkg string, f *descriptor.Field, reg *descriptor.Registry, opts opts.Options) (NamedFlowTyper, Dependencies, error) {
	// FieldMessage
	var fieldType FlowTyper = newSimpleType("any", opts)
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
		fieldType = newSimpleType("number", opts)
	case pbdescriptor.FieldDescriptorProto_TYPE_BOOL:
		fieldType = newSimpleType("boolean", opts)
	case pbdescriptor.FieldDescriptorProto_TYPE_STRING:
		fieldType = newSimpleType("string", opts)
	case pbdescriptor.FieldDescriptorProto_TYPE_GROUP:
		fieldType = newSimpleType("any", opts) // , required?
	case pbdescriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: should resolve type name relative to this type
		ft, err := reg.LookupMsg("", f.GetTypeName())
		if err != nil {
			return nil, nil, err
		}
		if flowType, present := knownTypeMap[ft.FQMN()]; present {
			fieldType = newSimpleType(flowType, opts)
		} else {
			fieldType = newMessageFlowType(cfg.messageTypeName(ft), opts)
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
		fieldType = newSimpleType("string", opts) // could be more correct
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
			fieldType = newSimpleType(name, opts)
		}
	}
	if f.GetLabel() == pbdescriptor.FieldDescriptorProto_LABEL_REPEATED {
		fieldType = newRepeatedFlowType(fieldType, opts)
	}
	return &namedType{FlowTyper: fieldType, name: f.GetName(), opts: opts}, deps, nil
}

func getFieldOptionsIfAny(field *pbdescriptor.FieldDescriptorProto) opts.Options {
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, opts.E_Field)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("issue getting field options: %v", err))
			return opts.Options{}
		}
		if o := v.(*opts.Options); o != nil {
			return *o
		}
	}
	return opts.Options{}
}

func (cfg GeneratorOptions) messageToFlowType(m *descriptor.Message, reg *descriptor.Registry) (FlowTyper, Dependencies, error) {
	deps := Dependencies{}
	t := &objectFlowType{
		Fields:  []NamedFlowTyper{},
		Options: cfg,
	}
	for _, f := range m.Fields {
		opts := cfg.ProtoOptions
		fieldOpts := getFieldOptionsIfAny(f.FieldDescriptorProto)
		if fieldOpts.Required != nil {
			opts.Required = fieldOpts.Required
		}
		if fieldOpts.Nullable != nil {
			opts.Nullable = fieldOpts.Nullable
		}
		field, newDeps, err := cfg.fieldToType(m.File.GetPackage(), f, reg, opts)
		if err != nil {
			return nil, nil, err
		}
		mergeDeps(deps, newDeps)
		t.Fields = append(t.Fields, field)
	}
	return &namedType{FlowTyper: t, name: cfg.messageTypeName(m)}, deps, nil
}

func (cfg GeneratorOptions) enumTypeName(e *descriptor.Enum) string {
	name := strings.Replace(e.FQEN(), ".", "", -1)
	if !cfg.AlwaysQualifyTypes {
		if strings.HasPrefix(name, e.File.GoPkg.Name) {
			name = name[len(e.File.GoPkg.Name):]
		}
	}
	return name
}

func (cfg GeneratorOptions) messageTypeName(m *descriptor.Message) string {
	name := strings.Replace(m.FQMN(), ".", "", -1)
	if !cfg.AlwaysQualifyTypes {
		if strings.HasPrefix(name, m.File.GoPkg.Name) {
			name = name[len(m.File.GoPkg.Name):]
		}
	}
	return name
}

func (cfg GeneratorOptions) enumToFlowType(e *descriptor.Enum, reg *descriptor.Registry) (FlowTyper, Dependencies, error) {
	options := []string{}
	for _, v := range e.Value {
		if !cfg.EmitEnumZeros && v.GetNumber() == 0 {
			continue
		}
		options = append(options, fmt.Sprintf(`"%s"`, v.GetName()))
	}
	name := cfg.enumTypeName(e)
	return &namedType{
		FlowTyper: newSimpleType(strings.Join(options, " | "), cfg.ProtoOptions),
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

func generateFlowTypes(file *descriptor.File, registry *descriptor.Registry, options GeneratorOptions) (string, error) {
	if options.DumpJSON {
		m := &jsonpb.Marshaler{EmitDefaults: true, OrigName: true, Indent: "  "}
		m.Marshal(os.Stderr, file)
		time.Sleep(time.Second)
	}

	if file.Options != nil {
		v, err := proto.GetExtension(file.Options, opts.E_FieldDefaults)
		if err == nil {
			if o := v.(*opts.Options); o != nil {
				options.ProtoOptions = *o
			}
		} else {
			if err != proto.ErrMissingExtension {
				return "", errors.Wrap(err, "GetExtenstion")
			}
		}
	}

	deps := Dependencies{}
	result := []FlowTyper{}
	f, err := registry.LookupFile(file.GetName())
	if err != nil {
		return "", err
	}
	for _, enum := range f.Enums {
		t, newDeps, err := options.enumToFlowType(enum, registry)
		if err != nil {
			return "", err
		}
		mergeDeps(deps, newDeps)
		result = append(result, t)
	}
	for _, message := range f.Messages {
		t, newDeps, err := options.messageToFlowType(message, registry)
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
{{ range $package, $types := .Dependencies -}}
import type {
{{range $type, $true := $types}}  {{$type}},
{{end -}}
} from './{{ $package }}.js';
{{ end }}

{{range .Result}}export type {{.Name}} = {{.FlowType}};

{{end}}
`)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, struct {
		GeneratorOptions
		Dependencies Dependencies
		Result       []FlowTyper
	}{GeneratorOptions: options, Dependencies: deps, Result: result})
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
