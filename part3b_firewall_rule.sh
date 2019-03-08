#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Create firewall rule

# Constants - CHANGE ME!
readonly PROJECT='go-srv-demo'

# Create firewall rule to allow ingress traffic from port 80
time gcloud compute firewall-rules create gke-go-srv-demo-cluster-service-a \
  --project $PROJECT \
  --description 'Allow access to Service A on port 8000' \
  --direction INGRESS \
  --priority 1000 \
  --network default \
  --action ALLOW \
  --rules tcp:8000 \
  --source-ranges 0.0.0.0/0
