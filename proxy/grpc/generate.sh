#!/bin/bash

# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# go install github.com/fatih/gomodifytags@v1.13.0
# go install github.com/FZambia/gomodifytype@latest

protoc -I proxyproto proxyproto/proxy.proto --go_out=./proxyproto --go-grpc_out=./proxyproto

gomodifytype -file proxyproto/proxy.pb.go -all -w -from "[]byte" -to "Raw"

bash generate_tags.sh