#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Dockerhub

declare -a arr=("a" "b" "c" "a" "d" "e" "f" "g" "h")
# declare -a arr=("a")
declare tag="1.0.0"

for i in "${arr[@]}"
do
  docker push "garystafford/go-srv-$i:$tag"
done

docker push "garystafford/angular-observe:$tag"
