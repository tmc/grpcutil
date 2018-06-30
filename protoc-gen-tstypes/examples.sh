#!/bin/bash
set -euo pipefail
set -x

cd testdata
for e in $(ls ./*proto); do
    protoc -I. --tstypes_out=v=2:output/ "${e}"
done
