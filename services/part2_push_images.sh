#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Dockerhub

# readonly -a arr=(a b c a d e f g h rev-proxy)
readonly -a arr=(a b e)
readonly tag=1.5.0

for i in "${arr[@]}"
do
  docker push "garystafford/go-srv-$i:$tag"
done

# docker push "garystafford/angular-observe:$tag"
