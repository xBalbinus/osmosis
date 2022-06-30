set -euo pipefail
# Install buf and gogo tools, so that differences that arise from
# toolchain differences are also caught.
readonly tools="$(mktemp -d)"
export PATH="${PATH}:${tools}/bin"
export GOBIN="${tools}/bin"
go get github.com/gogo/protobuf/protoc-gen-gogofaster@latest
go get github.com/bufbuild/buf/cmd/buf

#Specificially ignore all differences in go.mod / go.sum.
echo "Generating Protobuf files"
pwd
go run github.com/bufbuild/buf/cmd/buf generate
mv ./proto/osmosis/abci/types.pb.go ./abci/types/

if ! git diff --stat --exit-code ; then
    echo ">> ERROR:"
    echo ">>"
    echo ">> Protobuf generated code requires update (either tools or .proto files may have changed)."
    echo ">> Ensure your tools are up-to-date, re-run 'make proto-gen' and update this PR."
    echo ">>"
    exit 1
fi
