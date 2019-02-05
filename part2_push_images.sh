#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Push images to Dockerhub

declare -a arr=("a" "b" "c" "a" "d" "e" "f" "g" "h")

for i in "${arr[@]}"
do
  docker push "garystafford/go-srv-$i:1.0.0"
done
