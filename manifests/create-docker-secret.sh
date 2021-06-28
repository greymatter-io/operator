#!/bin/bash

kubectl create secret docker-registry docker.secret \
  -n gm-operator \
  --docker-server=docker.greymatter.io \
  --docker-username=$NEXUS_USER \
  --docker-password=$NEXUS_PASSWORD \
  --docker-email=$NEXUS_USER
