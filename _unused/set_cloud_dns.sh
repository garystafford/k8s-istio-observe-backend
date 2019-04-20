#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Update Cloud DNS A Records
# example: api.dev.example-api.com.

# Constants - CHANGE ME!
readonly CURRENT_PROJECT='go-srv-demo'
readonly DNS_PROJECT='gke-confluent-atlas'
readonly DOMAIN='example-api.com'
readonly ZONE='example-api'
readonly REGION='us-central1'
readonly TTL=120
readonly RECORDS=('dev' 'test')
readonly SUB_DOMAINS=('api' 'ui')

gcloud config set project ${CURRENT_PROJECT}

# Get the new LB IP address from the current project
readonly NEW_IP=$(gcloud compute forwarding-rules list \
    --filter "region:(${REGION})" \
  | awk 'NR==2 {print $3}')

gcloud config set project ${DNS_PROJECT}

# Make sure any old load balancers were removed
if [[ $(gcloud compute forwarding-rules list --filter "region:($REGION)" | wc -l | awk '{$1=$1};1') -gt 2 ]]; then
  echo "More than one load balancer detected, exiting script."
  exit 1
fi

# Get load balancer IP address from first record
readonly OLD_IP=$(gcloud dns record-sets list \
    --filter "name=api.${RECORDS[0]}.${DOMAIN}." --zone ${ZONE} \
  | awk 'NR==2 {print $4}')

echo "Old LB IP Address: ${OLD_IP}"
echo "New LB IP Address: ${NEW_IP}"

# Update DNS records
gcloud dns record-sets transaction start --zone ${ZONE}

for record in ${RECORDS[@]}; do
  for sd in ${SUB_DOMAINS[@]}; do
    echo "${sd}.${record}.${DOMAIN}."

    gcloud dns record-sets transaction remove \
        --name "${sd}.${record}.${DOMAIN}." --ttl ${TTL} \
        --type A --zone ${ZONE} "${OLD_IP}"

    gcloud dns record-sets transaction add \
        --name "${sd}.${record}.${DOMAIN}." --ttl ${TTL} \
        --type A --zone ${ZONE} "${NEW_IP}"
  done
done

gcloud dns record-sets transaction execute --zone ${ZONE}

gcloud config set project ${CURRENT_PROJECT}
