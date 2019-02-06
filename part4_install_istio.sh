#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Install Istio 1.0.5

readonly ISTIO_HOME="/Applications/istio-1.0.5"

helm repo add istio.io https://storage.googleapis.com/istio-prerelease/daily-build/master-latest-daily/charts
helm repo list

kubectl apply -f $ISTIO_HOME/install/kubernetes/helm/helm-service-account.yaml
helm init --service-account tiller

# Wait for Tiller pod to come up
# Error: could not find a ready tiller pod
sleep 15

helm install $ISTIO_HOME/install/kubernetes/helm/istio \
  --name istio \
  --namespace istio-system \
  --set global.mtls.enabled=true \
  --set grafana.enabled=true \
  --set kiali.enabled=true \
  --set prometheus.enabled=true \
  --set servicegraph.enabled=true \
  --set servicegraph.ingress.enabled=true \
  --set telemetry-gateway.grafanaEnabled=true \
  --set telemetry-gateway.prometheusEnabled=true \
  --set tracing.enabled=true \
  --set tracing.ingress.enabled=true \
  --set tracing.jaeger.ingress.enabled=true \
  --set tracing.provider=jaeger

# --set kiali.ingress.enabled=true \
# Doesn't work:
# Error: release istio failed: Ingress.extensions "kiali" is invalid:
# spec: Invalid value: []extensions.IngressRule(nil): either `backend` or `rules` must be specified

helm ls --all istio
helm list istio

# helm del --purge istio
# Doesn't work: Error: customresourcedefinitions.apiextensions.k8s.io "gateways.networking.istio.io" already exists
