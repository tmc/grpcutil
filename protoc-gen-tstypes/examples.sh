#!/bin/bash
set -euo pipefail
set -x

cd $(dirname $0)
PROTOC_GEN_TSTYPES_ROOT=$(pwd)

# We can set "PROTOBUF_ROOT" and "GOOGLEAPIS_ROOT" in `protoc-gen-tstypes/.env` so that we don't need to specify these
# environment variables every time we run this script. This file will be ignore by git.
[ -f "./.env" ] && source ./.env

# We need to clone https://github.com/protocolbuffers/protobuf and set the environment variable "PROTOBUF_ROOT" as the absolute path of it.
# This repository provide some .proto files we want like "google/protobuf/timestamp.proto".
echo "$PROTOBUF_ROOT"

# We need to clone https://github.com/googleapis/googleapis and set the environment variable "PROTOBUF_ROOT" as the absolute path of it.
# This repository provide some .proto files we want like "google/api/field_behavior.proto".
echo "$GOOGLEAPIS_ROOT"

cd testdata
rm -fr output/*
ds=(output/defaults output/int-enums output/camel-case-names output/outpattern-{1,2,3} output/wo-namespace output/async-iterators)

for proto_file in any.proto duration.proto empty.proto struct.proto timestamp.proto wrappers.proto; do
    ls -ln "$PROTOBUF_ROOT/src/google/protobuf/${proto_file}"
    ln -sf "$PROTOBUF_ROOT/src/google/protobuf/${proto_file}" "./${proto_file}"
done

mkdir -p ${ds[*]}
for e in ./*proto; do
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1:output/defaults/ "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,int_enums=true:output/int-enums/ "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,original_names=false:output/camel-case-names/ "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,outpattern={{.Dir}}/{{.BaseName}}.d.ts:output/outpattern-1/ "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out 'v=1,outpattern={{.Descriptor.GetPackage | replace "." "/"}}/{{.BaseName}}.d.ts:output/outpattern-2/' "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out 'v=1,outpattern={{.Dir}}/{{.BaseName}}pb.d.ts:output/outpattern-3/' "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,declare_namespace=false:output/wo-namespace/ "${e}"
    protoc -I. -I${PROTOC_GEN_TSTYPES_ROOT} -I${PROTOBUF_ROOT} -I${GOOGLEAPIS_ROOT} --tstypes_out=v=1,async_iterators=true:output/async-iterators/ "${e}"
done

cd $PROTOC_GEN_TSTYPES_ROOT

# install typescript if "node_modules" doesn't exist
[ ! -d "./node_modules" ] && npm install

if [ "${CHECK:-}" != "0" ]; then
    for d in ${ds[*]}; do
        set +e
        npx tsc --lib es2015,esnext.asynciterable --strict --pretty testdata/${d}/*ts
        set -e
    done
fi
