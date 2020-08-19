package gentstypes

// TODO: add nested messages support
// TODO: add nested enum support

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/jhump/protoreflect/desc"
	"github.com/tmc/grpcutil/protoc-gen-tstypes/opts"

	"google.golang.org/genproto/googleapis/api/annotations"
)

const indent = "    "

type MessageOptionsFunc = func(*desc.MessageDescriptor) MessageOptions
type FieldOptionsFunc = func(MessageOptions, *desc.FieldDescriptor) FieldOptions

type Parameters struct {
	AsyncIterators        bool
	DeclareNamespace      bool
	OutputNamePattern     string
	DumpRequestDescriptor bool
	EnumsAsInt            bool
	OriginalNames         bool
	Verbose               int
	Int64AsString         bool
	// TODO: allow template specification?

	MessageOptionsFunc MessageOptionsFunc
	FieldOptionsFunc   FieldOptionsFunc
}

type Generator struct {
	*bytes.Buffer
	indent       string
	Request      *plugin.CodeGeneratorRequest
	Response     *plugin.CodeGeneratorResponse
	usedPackages map[string]bool // Use this map to track which packages has been used in one TS file.
}

type OutputNameContext struct {
	Dir        string
	BaseName   string
	Descriptor *desc.FileDescriptor
	Request    *plugin.CodeGeneratorRequest
}

type MessageOptions struct {
	DefaultFieldOptions *FieldOptions
}

type FieldOptions struct {
	IsRequired bool
}

func New() *Generator {
	return &Generator{
		Buffer:   new(bytes.Buffer),
		Request:  new(plugin.CodeGeneratorRequest),
		Response: new(plugin.CodeGeneratorResponse),
	}
}

func (g *Generator) incIndent() {
	g.indent += indent
}

func (g *Generator) decIndent() {
	g.indent = string(g.indent[:len(g.indent)-len(indent)])
}

// WriteLine is used to add one line into the buffer
func (g *Generator) WriteLine(s string) {
	g.Buffer.WriteString(g.indent)
	g.Buffer.WriteString(s)
	g.Buffer.WriteString("\n")
}

// W is an alias for WriteLine
func (g *Generator) W(s string) {
	g.WriteLine(s)
}

func (g *Generator) writeComment(s string) {
	if s != "" {
		for _, line := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
			g.W(fmt.Sprintf("//%s", line))
		}
	}
}

var s = &spew.ConfigState{
	Indent:                  " ",
	DisableMethods:          true,
	SortKeys:                true,
	SpewKeys:                true,
	MaxDepth:                12,
	DisablePointerAddresses: true,
	DisableCapacities:       true,
}

func genName(r *plugin.CodeGeneratorRequest, f *desc.FileDescriptor, outPattern string) string {
	// TODO: consider using go_package if present?

	n := filepath.Base(f.GetName())
	if strings.HasSuffix(n, ".proto") {
		n = n[:len(n)-len(".proto")]
	}
	ctx := &OutputNameContext{
		Dir:        filepath.Dir(f.GetName()),
		BaseName:   n,
		Descriptor: f,
		Request:    r,
	}
	var t = template.Must(template.New("gentstypes/generator.go:genName").Funcs(sprig.FuncMap()).Parse(outPattern))
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, ctx); err != nil {
		log.Fatalln("issue rendering template:", err)
	}
	return buf.String()
}

func (g *Generator) GenerateAllFiles(params *Parameters) {
	files, err := desc.CreateFileDescriptors(g.Request.ProtoFile)
	if params.DumpRequestDescriptor {
		s.Fdump(os.Stderr, g.Request)
	}
	if err != nil {
		log.Fatal(err)
	}
	names := []string{}
	for _, fname := range g.Request.FileToGenerate {
		names = append(names, fname)
	}
	sort.Strings(names)
	for _, n := range names {
		f := files[n]
		g.W("// Code generated by protoc-gen-tstypes. DO NOT EDIT.\n")
		g.generate(f, params)
	}
}

func (g *Generator) generate(f *desc.FileDescriptor, params *Parameters) {
	g.usedPackages = make(map[string]bool)

	// TODO: consider best order
	ns := params.DeclareNamespace && f.GetPackage() != ""
	if ns {
		g.W(fmt.Sprintf("declare namespace %s {\n", f.GetPackage()))
		g.incIndent()
	}

	g.generateEnums(f.GetEnumTypes(), params)
	g.generateMessages(f.GetMessageTypes(), params)
	g.generateServices(f.GetServices(), params)

	if ns {
		g.decIndent()
		g.W("}\n")
	}

	str := g.Buffer.String()
	g.Buffer.Reset()

	g.generateDependencies(f, f.GetDependencies(), params)
	g.Buffer.WriteString(str)

	n := genName(g.Request, f, params.OutputNamePattern)
	if params.Verbose > 0 {
		fmt.Fprintln(os.Stderr, "generating", n)
	}
	g.Response.File = append(g.Response.File, &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(n),
		Content: proto.String(g.String()),
	})
	g.Buffer.Reset()
	g.usedPackages = make(map[string]bool)
}

func (g *Generator) generateDependencies(baseFile *desc.FileDescriptor, dependencies []*desc.FileDescriptor, params *Parameters) {
	for _, d := range dependencies {

		importLine := fmt.Sprintf(`import * as %s from "%s"`, formatImportModule(d), formatImportFile(baseFile, d))
		if used, ok := g.usedPackages[d.GetPackage()]; ok && used {
			g.W(importLine)
		} else {
			g.W(fmt.Sprintf(`// %s // imported but not used `, importLine))
		}
	}

	if len(dependencies) > 0 {
		g.W("") // add a new line after imports
	}

}

func (g *Generator) generateMessages(messages []*desc.MessageDescriptor, params *Parameters) {
	for _, m := range messages {
		g.generateMessage(m, params)
	}
}
func (g *Generator) generateEnums(enums []*desc.EnumDescriptor, params *Parameters) {
	for _, e := range enums {
		g.generateEnum(e, params)
	}
}
func (g *Generator) generateServices(services []*desc.ServiceDescriptor, params *Parameters) {
	for _, e := range services {
		g.generateService(e, params)
	}
}

func DefaultMessageOptionsFunc(m *desc.MessageDescriptor) MessageOptions {
	result := MessageOptions{}
	if o, err := proto.GetExtension(m.AsDescriptorProto().Options, opts.E_FieldDefaults); err == nil {
		if o, ok := o.(*opts.Options); ok {
			fieldRequiredDefault := o.GetRequired() || o.GetFieldBehavior() == annotations.FieldBehavior_REQUIRED
			result.DefaultFieldOptions = &FieldOptions{IsRequired: fieldRequiredDefault}
		}
	}
	return result
}

func DefaultFieldOptionsFunc(mOpts MessageOptions, f *desc.FieldDescriptor) FieldOptions {
	required := false
	if mOpts.DefaultFieldOptions != nil {
		required = mOpts.DefaultFieldOptions.IsRequired
	}
	e, err := proto.GetExtension(f.AsFieldDescriptorProto().Options, opts.E_Field)
	if err == nil {
		if e, ok := e.(*opts.Options); ok {
			required = e.GetRequired()
		}
	}
	if o, err := proto.GetExtension(f.AsFieldDescriptorProto().Options, annotations.E_FieldBehavior); err == nil {
		if opts, ok := o.([]annotations.FieldBehavior); ok {
			for _, opt := range opts {
				if opt == annotations.FieldBehavior_REQUIRED {
					required = true
				}
			}
		}
	}
	return FieldOptions{IsRequired: required}
}

/*
generateMessageInterface is used to generate TypeScript interface:

export interface MyMessage {
    id?: string;
}
*/
func (g *Generator) generateMessageInterface(m *desc.MessageDescriptor, params *Parameters) {
	name := m.GetName()

	mOpts := DefaultMessageOptionsFunc(m)
	if params.MessageOptionsFunc != nil {
		mOpts = params.MessageOptionsFunc(m)
	}

	g.writeComment(m.GetSourceInfo().GetLeadingComments())
	g.W(fmt.Sprintf("export interface %s {", name))
	for _, f := range m.GetFields() {
		name := f.GetName()
		if !params.OriginalNames {
			name = f.GetJSONName()
		}
		fOptsFn := DefaultFieldOptionsFunc
		if params.FieldOptionsFunc != nil {
			fOptsFn = params.FieldOptionsFunc
		}
		fOpts := fOptsFn(mOpts, f)
		required := fOpts.IsRequired

		suffix := ""
		if !required {
			suffix = "?"
		}

		g.incIndent()
		g.writeComment(f.GetSourceInfo().GetLeadingComments())
		g.decIndent()
		trailingComment := ""
		if comment := f.GetSourceInfo().GetTrailingComments(); comment != "" {
			trailingComment = " // " + strings.TrimSpace(comment)
		}
		g.W(fmt.Sprintf(indent+"%s%s: %s;%s", name, suffix, fieldType(f, params, g.usedPackages), trailingComment))
	}
	g.W("}\n")
}

/*
generateMessageNamespace is used to generate TypeScript namespace:

export namespace MyMessage {
	id?: string;
	child?: SubMessage;

	export interface SubMessage {
		id?: string;
	}

	export namespace SubMessage {
		...
	}
}
*/
func (g *Generator) generateMessageNamespace(m *desc.MessageDescriptor, params *Parameters) {
	name := m.GetName()

	nestedEnumTypes := m.GetNestedEnumTypes()
	nestedMessageTypes := m.GetNestedMessageTypes()

	// If there is not nested object, don't write the namespace
	if len(nestedEnumTypes) == 0 && len(nestedMessageTypes) == 0 {
		return
	}

	g.W(fmt.Sprintf("export namespace %s {", name))
	g.incIndent()

	for _, e := range nestedEnumTypes {
		g.generateEnum(e, params)
	}

	for _, m := range nestedMessageTypes {
		g.generateMessage(m, params)
	}

	g.decIndent()
	g.W("}\n")
}

func (g *Generator) generateMessage(m *desc.MessageDescriptor, params *Parameters) {
	g.generateMessageInterface(m, params)
	g.generateMessageNamespace(m, params)
}

func fieldType(f *desc.FieldDescriptor, params *Parameters, usedPackages map[string]bool) string {
	t := rawFieldType(f, params, usedPackages)
	if f.IsMap() {
		return fmt.Sprintf("{ [key: %s]: %s }", rawFieldType(f.GetMapKeyType(), params, usedPackages), rawFieldType(f.GetMapValueType(), params, usedPackages))
	}
	if f.IsRepeated() {
		return fmt.Sprintf("Array<%s>", t)
	}
	return t
}

func rawFieldType(f *desc.FieldDescriptor, params *Parameters, usedPackages map[string]bool) string {
	switch f.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return "number"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		if params.Int64AsString {
			return "string"
		} else {
			return "number"
		}
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "boolean"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		t := f.GetEnumType()
		if t.GetFile().GetPackage() != f.GetFile().GetPackage() {
			// this field is imported from the outside
			usedPackages[t.GetFile().GetPackage()] = true
			return formatImportModule(t.GetFile()) + "." + t.GetName()
		}
		return packageQualifiedName(t)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		t := f.GetMessageType()
		if t.GetFile().GetPackage() != f.GetFile().GetPackage() {
			// this field is imported from the outside
			usedPackages[t.GetFile().GetPackage()] = true
			return formatImportModule(t.GetFile()) + "." + t.GetName()
		}
		return packageQualifiedName(t)
	}
	return "any /*unknown*/"
}

func packageQualifiedName(e desc.Descriptor) string {
	name := e.GetName()
	var c desc.Descriptor
	for c = e.GetParent(); c.GetParent() != nil; c = c.GetParent() {
		name = fmt.Sprintf("%v.%v", c.GetName(), name)
	}
	return name
}

func (g *Generator) generateEnum(e *desc.EnumDescriptor, params *Parameters) {
	name := e.GetName()

	g.writeComment(e.GetSourceInfo().GetLeadingComments())
	g.W(fmt.Sprintf("export enum %s {", name))
	for _, v := range e.GetValues() {
		if params.EnumsAsInt {
			g.W(fmt.Sprintf("    %s = %v,", v.GetName(), v.GetNumber()))
		} else {
			g.W(fmt.Sprintf("    %s = \"%v\",", v.GetName(), v.GetName()))
		}
	}
	g.W("}")
}

func (g *Generator) generateService(service *desc.ServiceDescriptor, params *Parameters) {
	g.W(fmt.Sprintf("export interface %sService {", service.GetName()))
	g.incIndent()
	g.generateServiceMethods(service, params)
	g.decIndent()
	g.W(fmt.Sprintf("}"))
}

func (g *Generator) generateServiceMethods(service *desc.ServiceDescriptor, params *Parameters) {
	for _, m := range service.GetMethods() {
		g.generateServiceMethod(m, params)
	}
}
func (g *Generator) generateServiceMethod(method *desc.MethodDescriptor, params *Parameters) {
	i := method.GetInputType().GetName()
	o := method.GetOutputType().GetName()
	if params.AsyncIterators {
		if method.IsServerStreaming() {
			o = fmt.Sprintf("AsyncIterator<%s>", o)
		}
		if method.IsClientStreaming() {
			i = fmt.Sprintf("AsyncIterator<%s>", i)
		}
		g.W(fmt.Sprintf("%s: (r:%s) => %s;", method.GetName(), i, o))
	} else {
		ss, cs := method.IsServerStreaming(), method.IsClientStreaming()
		if !(ss || cs) {
			g.W(fmt.Sprintf("%s: (r:%s) => %s;", method.GetName(), i, o))
			return
		}
		if !cs {
			g.W(fmt.Sprintf("%s: (r:%s, cb:(a:{value: %s, done: boolean}) => void) => void;", method.GetName(), i, o))
			return
		}
		if !ss {
			g.W(fmt.Sprintf("%s: (r:() => {value: %s, done: boolean}) => %s;", method.GetName(), i, o))
			return
		}
		g.W(fmt.Sprintf("%s: (r:() => {value: %s, done: boolean}, cb:(a:{value: %s, done: boolean}) => void) => void;", method.GetName(), i, o))
	}
}
