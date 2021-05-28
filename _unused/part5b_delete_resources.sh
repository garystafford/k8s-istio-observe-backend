#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Delete Kubernetes resources

kubectl delete namespace dev test

helm delete --purge istio
helm delete --purge istio-init

kubectl get all -n dev
kubectl get all -n test
istioctl get all
