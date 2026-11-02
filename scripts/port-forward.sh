#!/bin/bash

killall kubectl 2>/dev/null

kubectl port-forward -n catalogio svc/catalogio-service 8080:8080 >/dev/null 2>&1 &

kubectl port-forward -n catalogio deployment/catalogio-app 8081:8081 >/dev/null 2>&1 &

kubectl port-forward -n catalogio svc/victoria-logs 9428:9428 >/dev/null 2>&1 &

kubectl port-forward -n catalogio svc/grafana 3000:3000 >/dev/null 2>&1 &

kubectl port-forward -n monitoring svc/vm-single-victoria-metrics-single-server 8428:8428 >/dev/null 2>&1 &

kubectl port-forward -n monitoring svc/alertmanager 9093:9093 >/dev/null 2>&1 &

kubectl port-forward -n monitoring svc/vmalert 8880:8880 >/dev/null 2>&1 &