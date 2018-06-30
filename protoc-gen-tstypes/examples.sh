#!/bin/bash
set -euo pipefail
set -x

cd testdata
rm -fr output/*
mkdir output/defaults output/int-enums output/outpattern output/wo-namespace
for e in $(ls ./*proto); do
    protoc -I. --tstypes_out=v=1:output/defaults/ "${e}"
    protoc -I. --tstypes_out=v=1,int_enums=true:output/int-enums/ "${e}"
    protoc -I. --tstypes_out=v=1,outpattern={{.Dir}}/{{.BaseName}}.d.ts:output/outpattern/ "${e}"
    protoc -I. --tstypes_out=v=1,declare_namespace=false:output/wo-namespace/ "${e}"
done
