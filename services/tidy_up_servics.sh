#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo
# date: 2021-07-07

readonly -a arr1=(a b c d e f g h)
for i in "${arr1[@]}"
do
  pushd "json-rest/service-$i" || exit
  go mod tidy -v
  popd || exit
done

readonly -a arr2=(a-grpc b-grpc c-grpc d-grpc e-grpc f-grpc g-grpc h-grpc rev-proxy-grpc)
for i in "${arr2[@]}"
do
  pushd "protobuf-grpc/service-$i" || exit
  go mod tidy -v
  popd || exit
done
