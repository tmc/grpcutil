// Program protoc-gen-tstypes generates TypeScript type declaration files from a Protocol Buffer file.
//
// Basic Example:
//  protoc -I. --tstypes_out=. simple.proto
//
// See examples.sh for more complex examples (output is in testdata/output)
//
// Parameters:
//  outpattern: control the output file paths.
//  int_enums: use ints instead of strings for enums (default false)
//  declare_namespace: declare namespace for the generated type (default true)
//  async_iterators: use async iterators for streaming endpoint types (default false)
package main
