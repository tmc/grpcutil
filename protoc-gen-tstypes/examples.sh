#!/bin/bash
set -euo pipefail
set -x

for e in $(ls ./testdata/example*proto); do
    protoc -I. --tstypes_out=v=2:testdata/output/ "${e}"
done
