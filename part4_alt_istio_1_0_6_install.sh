#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Install Istio 1.0.6

# set -ex

readonly ISTIO_HOME='/Applications/istio-1.0.6'

helm repo add istio.io https://storage.googleapis.com/istio-prerelease/daily-build/master-latest-daily/charts
helm repo list

kubectl apply -f ${ISTIO_HOME}/install/kubernetes/helm/helm-service-account.yaml
helm init --service-account tiller

# Wait for Tiller pod to come up
# Error: could not find a ready tiller pod
sleep 15

helm install ${ISTIO_HOME}/install/kubernetes/helm/istio \
  --name istio \
  --namespace istio-system \
  --set prometheus.enabled=true \
  --set grafana.enabled=true \
  --set kiali.enabled=true \
  --set tracing.enabled=true

kubectl apply --namespace istio-system -f ./resources/secrets/kiali.yaml

helm list istio
