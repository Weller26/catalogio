#!/bin/bash
set -e

kubectl apply -f k8s/00-namespace.yml

set -a
source .env
set +a

envsubst < k8s/01-secrets.yml.template | kubectl apply -f -

kubectl create configmap migrations-files --from-file=./migrations -n catalogio --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f k8s/