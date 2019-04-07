#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build protobuf files

protoc -I /usr/local/include -I. \
  -I $GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=plugins=grpc:. \
  greeting/greeting.proto

protoc -I /usr/local/include -I. \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:. \
  greeting/greeting.proto

protoc -I /usr/local/include -I. \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:. \
  greeting/greeting.proto
