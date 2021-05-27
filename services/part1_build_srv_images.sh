#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo
# date: 2021-05-24

readonly -a arr=(a b c d e f g h)
# readonly -a arr=(a)
readonly tag=1.6.5

for i in "${arr[@]}"
do
  cp -f Dockerfile "service-$i"
  pushd "service-$i" || exit
  docker build -t "garystafford/go-srv-$i:$tag" . --no-cache
  rm -rf Dockerfile
  popd || exit
done

docker image ls | grep 'garystafford/go-srv-'
