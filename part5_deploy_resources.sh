#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Deploy Kubernetes/Istio resources

# Constants - CHANGE ME!
readonly NAMESPACES=(dev test)
readonly SERVICES=(a b c d e f g h)


# Create Namespaces
kubectl apply -f ./resources/other/namespaces.yaml

for namespace in ${NAMESPACES[@]}; do
  # Enable automatic Istio sidecar injection
  kubectl label namespace $namespace istio-injection=enabled

  # Istio Gateway and three ServiceEntry resources
  kubectl apply -f ./resources/other/istio-gateway.yaml
  kubectl apply -n $namespace -f ./resources/config/go-srv-demo.yaml
  kubectl apply -n $namespace -f ./resources/services/rabbitmq.yaml
  kubectl apply -n $namespace -f ./resources/services/mongodb.yaml

  for service in ${SERVICES[@]}; do
    kubectl apply -n $namespace -f ./resources/services/service-$service.yaml
  done
done
