#!/bin/bash
set -e

# remove
rm dynamic-refresh-operator

# build
make

# create namespace
oc delete ns dynamic-refresh-operator
oc create ns dynamic-refresh-operator

# start the operator
#./dynamic-refresh-operator start --kubeconfig /home/ckyal/.kube/config --namespace dynamic-refresh-operator

./dynamic-refresh-operator start --kubeconfig /home/ckyal/.kube/config --namespace dynamic-refresh-operator --listen 0.0.0.0:8471