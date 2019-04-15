#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Docker Hub and GCR

#readonly -a arr=(a b c d e f g h rev-proxy)
readonly -a arr=(rev-proxy)
readonly tag=1.5.0

for i in ${arr[@]}
do
  docker push garystafford/go-srv-${i}:${tag}
  docker tag garystafford/go-srv-${i}:${tag} gcr.io/go-srv-demo/go-srv-${i}:${tag}
  docker push gcr.io/go-srv-demo/go-srv-${i}:${tag}
done

docker push garystafford/angular-observe:${tag}
docker tag garystafford/angular-observe:${tag} gcr.io/go-srv-demo/angular-observe:${tag}
docker push gcr.io/go-srv-demo/angular-observe:${tag}
