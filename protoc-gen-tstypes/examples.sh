#!/bin/bash
set -euo pipefail
set -x

cd testdata
rm -fr output/*
ds=(output/defaults output/int-enums output/original-names output/outpattern-{1,2,3} output/wo-namespace output/async-iterators)

# GOPATH src root relative to the testdata directory
gopath_root="../../../../../"

mkdir ${ds[*]}
for e in $(ls ./*proto); do
    protoc -I. -I${gopath_root} --tstypes_out=v=1:output/defaults/ "${e}"
    protoc -I. -I${gopath_root} --tstypes_out=v=1,int_enums=true:output/int-enums/ "${e}"
    protoc -I. -I${gopath_root} --tstypes_out=v=1,original_names=true:output/original-names/ "${e}"
    protoc -I. -I${gopath_root} --tstypes_out=v=1,outpattern={{.Dir}}/{{.BaseName}}.d.ts:output/outpattern-1/ "${e}"
    protoc -I. -I${gopath_root} --tstypes_out 'v=1,outpattern={{.Descriptor.GetPackage | replace "." "/"}}/{{.BaseName}}.d.ts:output/outpattern-2/' "${e}"
    protoc -I. -I${gopath_root} --tstypes_out 'v=1,outpattern={{.Dir}}/{{.BaseName}}pb.d.ts:output/outpattern-3/' "${e}"
    protoc -I. -I${gopath_root} --tstypes_out=v=1,declare_namespace=false:output/wo-namespace/ "${e}"
    protoc -I. -I${gopath_root} --tstypes_out=v=1,async_iterators=true:output/async-iterators/ "${e}"
done

if [ "${CHECK:-}" != "" ]; then
    for d in ${ds[*]}; do
        set +e
        tsc --lib es2015,esnext.asynciterable --strict --pretty ${d}/*ts
        set -e
    done
fi
