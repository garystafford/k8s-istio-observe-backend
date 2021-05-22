#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Build Go microservices for demo
# date: 2021-05-22

readonly -a arr=(a b c d e f g h)
# readonly -a arr=(a)

for i in "${arr[@]}"
do
  pushd "service-$i" || exit
  go mod init "github.com/garystafford/go-srv-$i"
  go mod tidy -v
  popd || exit
done
