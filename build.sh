pushd gitpb && \
protoc -I. --grpc-gateway_out=logtostderr=true:. --gofast_out=plugins=grpc:. *.proto && \
popd && \
go fmt ./... && \
go build
