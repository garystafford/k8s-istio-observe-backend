#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Dockerhub
# date: 2021-05-24

readonly -a arr=(a b c d e f g h)
# readonly -a arr=(a)
readonly tag=1.6.0

for i in "${arr[@]}"
do
  docker push "docker.io/garystafford/go-srv-$i:$tag"
done

docker push "docker.io/garystafford/angular-observe:$tag"
