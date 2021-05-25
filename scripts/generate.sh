#!/usr/bin/env bash

CODEGEN_DIR="$GOPATH/src/github.com/kubernetes/code-generator"
SRC_PATH=github.com/bcmendoza/gm-operator/apis
DEST_PATH=github.com/bcmendoza/gm-operator/client
REPO_PATH="$GOPATH/src/github.com/bcmendoza/gm-operator"
TAGS=install:v1

$CODEGEN_DIR/generate-groups.sh all $DEST_PATH $SRC_PATH $TAGS \
  --go-header-file "$REPO_PATH/scripts/license.go.txt"
