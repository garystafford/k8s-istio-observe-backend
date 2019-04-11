#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo

# readonly -a arr=(a b c a d e f g h rev-proxy)
readonly -a arr=(rev-proxy)
readonly tag=1.5.0

for i in "${arr[@]}"
do
  cp -f Dockerfile "service-$i"
  pushd "service-$i"
  docker build -t "garystafford/go-srv-$i:$tag" . --no-cache
  rm -rf Dockerfile
  popd
done

docker image ls | grep 'garystafford/go-srv-'
