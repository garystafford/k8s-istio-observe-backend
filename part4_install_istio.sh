#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Install Istio 1.0.5

readonly ISTIO_HOME="/Applications/istio-1.0.5"

helm repo add istio.io https://storage.googleapis.com/istio-prerelease/daily-build/master-latest-daily/charts

kubectl apply -f $ISTIO_HOME/install/kubernetes/helm/helm-service-account.yaml
helm init --service-account tiller
sleep 15
helm install $ISTIO_HOME/install/kubernetes/helm/istio \
  --name istio \
  --namespace istio-system \
  --set mtls.enable=true \
  --set grafana.enable=true \
  --set prometheus.enable=true \
  --set servicegraph.enable=true \
  --set tracing.enabled=true \
  --set tracing.provider=zipkin \
  --set kiali.enable=true
  #  \
  # --set telemetry-gateway.grafanaEnabled=true \
  # --set telemetry-gateway.prometheusEnabled=true \

helm ls --all istio
helm list istio

# helm upgrade \
#   --install istio \
#   --namespace istio-system $ISTIO_HOME/install/kubernetes/helm/istio
