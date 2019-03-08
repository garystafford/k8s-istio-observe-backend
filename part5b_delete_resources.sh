#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Delete Kubernetes resources

kubectl delete namespace dev test
istioctl delete gateway demo-gateway
istioctl delete serviceentry cloudamqp-external-mesh mongdb-atlas-external-mesh
istioctl delete virtualservice service-a-dev service-a-test angular-ui-dev angular-ui-test

kubectl get all -n dev
kubectl get all -n test
istioctl get all
