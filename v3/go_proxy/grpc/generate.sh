#!/bin/bash

# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc -I proxyproto proxyproto/proxy.proto --go_out=./proxyproto --go-grpc_out=./proxyproto