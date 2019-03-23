#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Dockerhub

readonly -a arr=(a b c a d e f g h)
# readonly -a arr=(a)
readonly tag=1.4.0

for i in "${arr[@]}"
do
  docker push "garystafford/go-srv-$i:$tag"
done

docker push "garystafford/angular-observe:$tag"
