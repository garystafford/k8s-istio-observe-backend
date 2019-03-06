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
  kubectl delete -f ./resources/other/istio-gateway.yaml
  kubectl delete -f ../golang-srv-demo-secrets/other/external-mesh-mongodb-atlas.yaml
  kubectl delete -f ../golang-srv-demo-secrets/other/external-mesh-cloudamqp.yaml
  kubectl delete -n $namespace -f ../golang-srv-demo-secrets/secret/go-srv-demo.yaml

  for service in ${SERVICES[@]}; do
    kubectl delete -n $namespace -f ./resources/services/service-$service.yaml
  done
  kubectl delete -n $namespace -f ./resources/services/angular-ui.yaml
done
