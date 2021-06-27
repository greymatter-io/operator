#!/bin/bash

# When adding this to the CLI, this should read from env vars or config file and send the following:
# apiVersion: v1
# kind: Secret
# metadata:
#   name: docker.secret
#   namespace: gm-operator
# type: kubernetes.io/dockerconfigjson
# data:
#   .dockerconfigjson: (BASE64 ENCODED STRING)

# The base64 encoded string should be created from the following JSON defined in the CLI:
# {
#   "auths": {
#     "docker.greymatter.io": {
#       "username": "$NEXUS_USER",
#       "password": "$NEXUS_PASSWORD",
#       "email": "$NEXUS_USER"
#     }
#   }
# }

kubectl create secret docker-registry docker.secret --dry-run=true \
  -n gm-operator \
  --docker-server=docker.greymatter.io \
  --docker-username=$NEXUS_USER \
  --docker-password=$NEXUS_PASSWORD \
  --docker-email=$NEXUS_USER
