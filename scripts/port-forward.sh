#!/bin/bash

killall kubectl 2>/dev/null
kubectl port-forward -n catalogio svc/catalogio-service 8080:8080 >/dev/null 2>&1 &
kubectl port-forward -n catalogio svc/victoria-logs 9428:9428 >/dev/null 2>&1 &
kubectl port-forward -n catalogio svc/grafana 3000:3000 >/dev/null 2>&1 &