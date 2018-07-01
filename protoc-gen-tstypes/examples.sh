#!/bin/bash
set -euo pipefail
set -x

cd testdata
rm -fr output/*
mkdir output/defaults output/int-enums output/outpattern-{1,2,3} output/wo-namespace output/async-iterators
for e in $(ls ./*proto); do
    protoc -I. --tstypes_out=v=1:output/defaults/ "${e}"
    protoc -I. --tstypes_out=v=1,int_enums=true:output/int-enums/ "${e}"
    protoc -I. --tstypes_out=v=1,outpattern={{.Dir}}/{{.BaseName}}.d.ts:output/outpattern-1/ "${e}"
    protoc -I. --tstypes_out 'v=1,outpattern={{.Descriptor.GetPackage | replace "." "/"}}/{{.BaseName}}.d.ts:output/outpattern-2/' "${e}"
    protoc -I. --tstypes_out 'v=1,outpattern={{.Dir}}/{{.BaseName}}pb.d.ts:output/outpattern-3/' "${e}"
    protoc -I. --tstypes_out=v=1,declare_namespace=false:output/wo-namespace/ "${e}"
    protoc -I. --tstypes_out=v=1,async_iterators=true:output/async-iterators/ "${e}"
done
