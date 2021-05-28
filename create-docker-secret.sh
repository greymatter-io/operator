#!/bin/bash

kubectl create secret docker-registry docker.secret \
  --docker-server=docker.greymatter.io \
  --docker-username=$NEXUS_USER \
  --docker-password=$NEXUS_PASSWORD \
  --docker-email=$NEXUS_USER
