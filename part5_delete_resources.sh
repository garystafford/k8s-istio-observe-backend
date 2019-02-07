#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Delete Kubernetes resources

# Constants - CHANGE ME!
readonly NAMESPACES=(dev test)
readonly SERVICES=(a b c d e f g h)

for namespace in ${NAMESPACES[@]}; do
  kubectl delete -n $namespace -f ./resources/services/rabbitmq.yaml
  kubectl delete -n $namespace -f ./resources/services/mongodb.yaml

  for service in ${SERVICES[@]}; do
    kubectl delete -n $namespace -f ./resources/services/service-$service.yaml
  done
done
