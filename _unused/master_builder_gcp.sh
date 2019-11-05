#!/bin/bash
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Do it all on Google Cloud

set -ex

time bash part3_create_gke_cluster.sh
export ISTIO_HOME="/home/gary_a_stafford/istio-1.1.2"
time bash part4_install_istio.sh
echo 'Waiting 30 seconds...'
sleep 30
time bash part5_deploy_resources.sh
time bash _unused/set_cloud_dns.sh