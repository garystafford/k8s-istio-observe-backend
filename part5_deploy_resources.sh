#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Deploy Kubernetes/Istio resources

# Constants - CHANGE ME!
readonly NAMESPACE='dev'

# Create Namespaces
kubectl apply -f ./resources/other/namespaces.yaml

# Enable automatic Istio sidecar injection
kubectl label namespace $NAMESPACE istio-injection=enabled

# Istio Gateway and three ServiceEntry resources
kubectl apply -f ./resources/other/istio-gateway.yaml
kubectl apply -n $NAMESPACE -f ./resources/config/go-srv-demo.yaml
kubectl apply -n $NAMESPACE -f ./resources/services/rabbitmq.yaml
kubectl apply -n $NAMESPACE -f ./resources/services/mongodb.yaml
kubectl apply -n $NAMESPACE -f ./resources/services/service-a.yaml
