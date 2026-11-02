#!/bin/bash

docker build -t catalogio-app:1.0 .
minikube image load catalogio-app:1.0

./scripts/apply-manifests.sh

kubectl rollout restart deployment/catalogio-app -n catalogio
kubectl rollout restart deployment/vmalert -n monitoring
kubectl rollout restart deployment/alertmanager -n monitoring

sleep 10

./scripts/port-forward.sh