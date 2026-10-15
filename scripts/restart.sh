#!/bin/bash

docker build -t catalogio-app:1.0 .
minikube image load catalogio-app:1.0
kubectl rollout restart deployment/catalogio-app -n catalogio

sleep 10

./scripts/port-forward.sh