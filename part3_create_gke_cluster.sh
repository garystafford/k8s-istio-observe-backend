#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Create 3-node GKE cluster

# Constants - CHANGE ME!
readonly PROJECT='go-srv-demo'
readonly CLUSTER='go-srv-demo-cluster'
readonly REGION='us-east1'
readonly MASTER_AUTH_NETS='72.231.208.0/24'
readonly NAMESPACE='dev'

# yes | gcloud components update
# gcloud init # set new project

# Build a 3-node, single-region, multi-zone GKE cluster
time gcloud beta container \
  --project $PROJECT clusters create $CLUSTER \
  --region $REGION \
  --no-enable-basic-auth \
  --no-issue-client-certificate \
  --cluster-version "1.11.6-gke.6" \
  --machine-type "n1-standard-1" \
  --image-type "COS" \
  --disk-type "pd-standard" \
  --disk-size "100" \
  --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" \
  --num-nodes "1" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  --enable-master-authorized-networks \
  --master-authorized-networks $MASTER_AUTH_NETS \
  --network "projects/${PROJECT}/global/networks/default" \
  --subnetwork "projects/${PROJECT}/regions/${REGION}/subnetworks/default" \
  --default-max-pods-per-node "110" \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing \
  --metadata disable-legacy-endpoints=true \
  --enable-autoupgrade \
  --enable-autorepair

# Get cluster credentials
gcloud container clusters get-credentials $CLUSTER \
  --region $REGION --project $PROJECT

kubectl config current-context
