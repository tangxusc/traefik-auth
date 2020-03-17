#!/usr/bin/env bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
docker build . -t auth:v1
rm -rf main
kubectl apply -f traefik-deploy.yaml -n traefik
kubectl apply -f test.yaml -n traefik
