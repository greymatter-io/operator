#!/bin/bash

kubectl create namespace gm-operator-system

kubectl create secret docker-registry docker.secret \
  -n gm-operator-system \
  --docker-server=docker.greymatter.io \
  --docker-username=$NEXUS_USER \
  --docker-password=$NEXUS_PASSWORD \
  --docker-email=$NEXUS_USER

kubectl create secret docker-registry docker.secret \
  --docker-server=docker.greymatter.io \
  --docker-username=$NEXUS_USER \
  --docker-password=$NEXUS_PASSWORD \
  --docker-email=$NEXUS_USER
