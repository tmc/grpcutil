package genflowtypes

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	pbdescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/mwitkow/go-proto-validators"
)

// Flower is a flow language type
type Flower interface {
	FlowType() string
	IsRequired() bool
}

// NamedFlower is a Flower with a name
type NamedFlower interface {
	Flower
	Name() string
}

func newSimpleType(typeString string, required bool) *simpleType {
	return &simpleType{typeString, required}
}

type simpleType struct {
	typeString string
	required   bool
}

func (s simpleType) FlowType() string { return s.typeString }
func (s simpleType) IsRequired() bool { return s.required }

type messageType struct {
	*simpleType
}

func newMessageFlowType(typeName string, required bool) *messageType {
	s := newSimpleType(typeName, required)
	return &messageType{s}
}

type repeatedType struct {
	Flower
	required bool
}

func newRepeatedFlowType(underlying Flower, required bool) *repeatedType {
	return &repeatedType{Flower: underlying, required: required}
}

func (r repeatedType) FlowType() string { return fmt.Sprintf("Array<%s>", r.Flower.FlowType()) }

type namedType struct {
	Flower
	name     string
	required bool
}

func (t *namedType) Name() string {
	return t.name
}

type objectFlowType struct {
	Fields   []NamedFlower
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

func (cfg Options) fqmnToType(fqmn string, registry *descriptor.Registry) (Flower, error) {
	m, err := registry.LookupMsg("", fqmn)
	if err != nil {
		return nil, err
	}
	return cfg.messageToFlowType(m, registry)
}

func (cfg Options) fieldToType(f *descriptor.Field, reg *descriptor.Registry, required bool) (NamedFlower, error) {
	// FieldMessage
	var fieldType Flower = newSimpleType("any", required)
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
			return nil, err
		}
		fieldType = newMessageFlowType(cfg.messageTypeName(ft), required)
	case pbdescriptor.FieldDescriptorProto_TYPE_BYTES:
		fieldType = newSimpleType("string", required) // could be more correct
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
			fieldType = newSimpleType(name, required)
		}
	}
	if f.GetLabel() == pbdescriptor.FieldDescriptorProto_LABEL_REPEATED {
		fieldType = newRepeatedFlowType(fieldType, required)
	}
	return &namedType{Flower: fieldType, name: f.GetName(), required: required}, nil
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

func (cfg Options) messageToFlowType(m *descriptor.Message, reg *descriptor.Registry) (Flower, error) {
	t := &objectFlowType{
		Fields:  []NamedFlower{},
		Options: cfg,
	}
	for _, f := range m.Fields {
		required := false
		if validatorOptions := getFieldValidatorIfAny(f.FieldDescriptorProto); validatorOptions != nil {
			required = *validatorOptions.MsgExists
		}
		field, err := cfg.fieldToType(f, reg, required)
		if err != nil {
			return nil, err
		}
		t.Fields = append(t.Fields, field)
	}
	return &namedType{Flower: t, name: cfg.messageTypeName(m)}, nil
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

func (cfg Options) enumToFlowType(e *descriptor.Enum, reg *descriptor.Registry) (Flower, error) {
	options := []string{}
	for _, v := range e.Value {
		options = append(options, fmt.Sprintf(`"%s"`, v.GetName()))
	}
	name := cfg.enumTypeName(e)
	return &namedType{
		Flower: newSimpleType(strings.Join(options, " | "), false),
		name:   name,
	}, nil
}

func generateFlowTypes(file *descriptor.File, registry *descriptor.Registry, opts Options) (string, error) {
	result := []Flower{}
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
	tmpl, err := template.New("").Parse(`/* @flow */
/* eslint-disable */
// Code generated by protoc-gen-flowtypes DO NOT EDIT.

{{range .}}export type {{.Name}} = {{.FlowType}};

{{end}}
`)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, result)
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
