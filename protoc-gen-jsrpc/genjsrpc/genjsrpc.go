package genjsrpc

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/asaskevich/govalidator"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

func removePackage(s string) string {
	sp := strings.Split(s, ".")
	return sp[len(sp)-1]
}

func (cfg GeneratorOptions) methodToRPC(name string, m *descriptor.Method) (string, []string) {
	types := []string{}
	methodName := strings.ToLower(*m.Name)[0:1] + (*m.Name)[1:]
	inputType := removePackage(*m.InputType)
	outputType := removePackage(*m.OutputType)
	types = append(types, inputType, outputType)
	prefix := strings.ToUpper(name)
	actionName := prefix + "_" + strings.ToUpper(govalidator.CamelCaseToUnderscore(methodName))
	s := `export const ` + methodName + ` = (req: ` + inputType + `) => (dispatch: (action: any) => void) => {
  dispatch({
    type: '` + actionName + `_REQUEST',
  })
  client().` + methodName + `(req, (error: ?Error, response: ?` + outputType + `) => {
    if (error) {
      dispatch({
        type: 'ERROR',
        payload: {error, action: '` + actionName + `'},
      })
      return
    }
    dispatch({
      type: '` + actionName + `_RESPONSE',
      payload: response,
    })
  })
}`
	return s, types
}

func (cfg GeneratorOptions) serviceToRPC(s *descriptor.Service, reg *descriptor.Registry) ([]string, []string, error) {
	types := []string{}
	result := []string{}
	for _, m := range s.Methods {
		glog.V(1).Infof("Method: %s", spew.Sdump(m))
		s, t := cfg.methodToRPC(*s.Name, m)
		types = append(types, t...)
		result = append(result, s)
	}
	return result, types, nil
}

func generateJSRPC(file *descriptor.File, registry *descriptor.Registry, options GeneratorOptions) (string, error) {
	types := []string{}
	result := []string{}
	f, err := registry.LookupFile(file.GetName())
	if err != nil {
		return "", err
	}
	for _, service := range f.Services {
		s, t, err := options.serviceToRPC(service, registry)
		if err != nil {
			return "", err
		}
		types = append(types, t...)
		result = append(result, s...)
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("").Parse(`/* eslint-disable */
// Code generated by protoc-gen-jsrpc DO NOT EDIT.
// InputID: {{.InputID}}
import {client} from './client'

import type {{"{"}}{{.FlowTypes}}{{"}"}} from './types'

{{.Result}}
`)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, struct {
		GeneratorOptions
		FlowTypes string
		Result    string
	}{GeneratorOptions: options, FlowTypes: strings.Join(types, ","), Result: strings.Join(result, "\n\n")})
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
