#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Create GKE cluster
# date: 2021-05-24

# Constants - CHANGE ME!
readonly PROJECT="go-srv-demo"
readonly CLUSTER="go-srv-demo-cluster"
readonly ZONE="us-east4-a"
readonly GKE_VERSION="11.19.9-gke.1400"
readonly MACHINE_TYPE="e2-medium"
readonly SERVICE_ACCOUNT="go-srv-demo@appspot.gserviceaccount.com"

# install gcloud cli
# https://formulae.brew.sh/cask/google-cloud-sdk
# brew install --cask google-cloud-sdk
# source "$(brew --prefix)/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/path.zsh.inc"
# source "$(brew --prefix)/Caskroom/google-cloud-sdk/latest/google-cloud-sdk/completion.zsh.inc"
# yes | gcloud components update

# gcloud init # set new project

# Build a 3-node, single-region, multi-zone GKE cluster
#gcloud beta container --project "go-srv-demo" clusters create "go-srv-demo-cluster" --zone "us-east4-a" --no-enable-basic-auth --cluster-version "1.19.9-gke.1400" --release-channel "regular" --machine-type "e2-medium" --image-type "COS_CONTAINERD" --disk-type "pd-standard" --disk-size "100" --metadata disable-legacy-endpoints=true --service-account "go-srv-demo@appspot.gserviceaccount.com" --num-nodes "3" --enable-stackdriver-kubernetes --enable-ip-alias --network "projects/go-srv-demo/global/networks/default" --subnetwork "projects/go-srv-demo/regions/us-east4/subnetworks/default" --no-enable-intra-node-visibility --default-max-pods-per-node "110" --enable-autoscaling --min-nodes "0" --max-nodes "3" --enable-dataplane-v2 --no-enable-master-authorized-networks --addons HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver --enable-autoupgrade --enable-autorepair --max-surge-upgrade 1 --max-unavailable-upgrade 0 --enable-autoprovisioning --min-cpu 2 --max-cpu 6 --min-memory 8 --max-memory 24 --autoprovisioning-locations=us-east4-a,us-east4-b,us-east4-c --autoprovisioning-service-account=go-srv-demo@appspot.gserviceaccount.com --enable-autoprovisioning-autorepair --enable-autoprovisioning-autoupgrade --autoprovisioning-max-surge-upgrade 1 --autoprovisioning-max-unavailable-upgrade 0 --enable-vertical-pod-autoscaling --enable-shielded-nodes --node-locations "us-east4-a","us-east4-b","us-east4-c"

gcloud beta container \
  --project "${PROJECT}" clusters create "${CLUSTER}" \
  --zone "${ZONE}" \
  --no-enable-basic-auth \
  --cluster-version "${CLUSTER}" \
  --release-channel "regular" \
  --machine-type "${MACHINE_TYPE}" \
  --image-type "COS_CONTAINERD" \
  --disk-type "pd-standard" \
  --disk-size "100" \
  --metadata disable-legacy-endpoints=true \
  --service-account "${SERVICE_ACCOUNT}" \
  --num-nodes "1" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  --network "projects/${PROJECT}/global/networks/default" \
  --subnetwork "projects/${PROJECT}/regions/${ZONE}/subnetworks/default" \
  --no-enable-intra-node-visibility \
  --default-max-pods-per-node "110" \
  --enable-autoscaling \
  --min-nodes "0" --max-nodes "3" \
  --enable-dataplane-v2 \
  --no-enable-master-authorized-networks \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing,Istio,GcePersistentDiskCsiDriver \
  --istio-config auth=MTLS_PERMISSIVE \
  --enable-autoupgrade \
  --enable-autorepair \
  --max-surge-upgrade 1 \
  --max-unavailable-upgrade 0 \
  --enable-autoprovisioning \
  --min-cpu 2 --max-cpu 6 \
  --min-memory 8 --max-memory 24 \
  --autoprovisioning-locations=us-east4-a,us-east4-b,us-east4-c \
  --autoprovisioning-service-account=${SERVICE_ACCOUNT} \
  --enable-autoprovisioning-autorepair \
  --enable-autoprovisioning-autoupgrade \
  --autoprovisioning-max-surge-upgrade 1 \
  --autoprovisioning-max-unavailable-upgrade 0 \
  --enable-vertical-pod-autoscaling \
  --enable-shielded-nodes \
  --node-locations "us-east4-a","us-east4-b","us-east4-c"

# Get cluster credentials
gcloud container clusters get-credentials "${CLUSTER}" \
  --region "${ZONE}" --project "${PROJECT}"

kubectl config current-context

#yes | gcloud container clusters delete go-srv-demo-cluster \
#  --region "${ZONE}" --project "${PROJECT}"