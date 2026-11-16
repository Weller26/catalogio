#!/bin/bash

set -e

minikube start -p minikube
minikube update-context -p minikube
kubectl config use-context minikube
kubectl --context=minikube --request-timeout=30s get --raw=/readyz

docker build \
    --build-arg APP_VERSION=v1.0.0 \
    --build-arg VCS_REF="$(git rev-parse --short HEAD)" \
    -t catalogio-app:1.0 .
minikube -p minikube image load catalogio-app:1.0

./scripts/apply-manifests.sh

kubectl rollout restart deployment/catalogio-app -n catalogio
kubectl rollout restart deployment/vmalert -n monitoring
kubectl rollout restart deployment/alertmanager -n monitoring

sleep 10

./scripts/port-forward.sh
