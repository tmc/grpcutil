#!/bin/bash
set -euo pipefail
set -x

cd testdata
rm -fr output/*
ds=(output/defaults output/int-enums output/camel-case-names output/outpattern-{1,2,3} output/wo-namespace output/async-iterators)

# GOPATH src root relative to the testdata directory
GOPATH_ROOT="../../../../../"
GOOGLEAPIS_ROOT="${GOPATH_ROOT}/github.com/googleapis/googleapis"

mkdir -p ${ds[*]}
for e in ./*proto; do
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1:output/defaults/ "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,int_enums=true:output/int-enums/ "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,original_names=false:output/camel-case-names/ "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,outpattern={{.Dir}}/{{.BaseName}}.d.ts:output/outpattern-1/ "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out 'v=1,outpattern={{.Descriptor.GetPackage | replace "." "/"}}/{{.BaseName}}.d.ts:output/outpattern-2/' "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out 'v=1,outpattern={{.Dir}}/{{.BaseName}}pb.d.ts:output/outpattern-3/' "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,declare_namespace=false:output/wo-namespace/ "${e}"
    protoc -I. -I${GOPATH_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,async_iterators=true:output/async-iterators/ "${e}"
done

if [ "${CHECK:-}" != "0" ]; then
    for d in ${ds[*]}; do
        set +e
        npx typescript --lib es2015,esnext.asynciterable --strict --pretty ${d}/*ts
        set -e
    done
fi
