#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo

readonly -a arr=(a b c d e f g h rev-proxy)
# readonly -a arr=(rev-proxy)
readonly tag=1.6.0-grpc

for srv in "${arr[@]}"
do
  cp -f Dockerfile service-"${srv}"
  pushd service-"${srv}" || exit
  docker build -t garystafford/go-srv-"${srv}":${tag} . --no-cache
  rm -rf Dockerfile
  popd
done

docker image ls | grep 'garystafford/go-srv-'
