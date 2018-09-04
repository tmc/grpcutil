// protoc-gen-tstypes generates TypeScript type declaration files from Protocol Buffer files. It aims to match the structure of the canonical JSON encoding for Protocol Buffers version 3.
//
// Example input:
//
// https://github.com/tmc/grpcutil/blob/master/protoc-gen-tstypes/testdata/route_guide.proto
//
// Example output:
//
// https://github.com/tmc/grpcutil/blob/master/protoc-gen-tstypes/testdata/output/defaults/routeguide.route_guide.d.ts
//
// Installation
//
// protoc-gen-tstypes is written in go. Assuming you have a working go installation (see https://golang.org/doc/install) you can fetch and build this project by executing:
//  go get github.com/tmc/grpcutil/protoc-gen-tstypes
// This will place the protoc-gen-tstypes binary in the $GOPATH/bin directory. If GOPATH has not been set it defaults to $HOME/go. The $GOPATH/bin directory should be on your PATH, see the install reference above for details.
//
// Usage
//
// Typical use will be via a protoc execution, a very simple example is:
//  protoc -I. --tstypes_out=. route_guide.proto
//
// See examples.sh for more complex examples (output is in testdata/output)
//
// Options
//
// The following options are available:
//  declare_namespace: declare namespace for the generated type (default true)
//  original_names: use original field names, otherwise use lowerCamelCase (default false)
//  int_enums: use ints instead of strings for enums (default false)
//  outpattern: control the output file paths.
//  async_iterators: use async iterators for streaming endpoint types (default false)
//  int64_string: use string representation for 64 bit numbers (default false)
// An example of running with a custom option set:
//  protoc -I. --tstypes_out=original_names=true,async_iterators=true:. route_guide.proto
//
// examples.sh contains more complex examples and generated output can be seen at https://github.com/tmc/grpcutil/blob/master/protoc-gen-tstypes/testdata/output
//
package main
