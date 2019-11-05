#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Docker Hub and GCR

readonly -a arr=(a b c d e f g h rev-proxy)
#readonly -a arr=(rev-proxy)
readonly tag=1.6.0-grpc

for srv in "${arr[@]}"
do
  docker push garystafford/go-srv-"${srv}":${tag}
done

docker push garystafford/angular-observe:${tag}