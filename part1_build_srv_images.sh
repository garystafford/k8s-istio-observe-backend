#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo

readonly -a arr=(a b c a d e f g h)
# declare -a arr=(a)
readonly tag=1.1.0

for i in "${arr[@]}"
do
  cp -f Dockerfile "service-$i"
  pushd "service-$i"
  docker build -t "garystafford/go-srv-$i:$tag" . --no-cache
  rm -rf Dockerfile
  popd
done

docker image ls | grep 'garystafford/go-srv-'

# docker stack rm golang-demo
# sleep 5
# docker network create -d overlay --attachable golang-demo
# docker stack deploy -c stack.yml golang-demo
