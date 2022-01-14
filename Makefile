# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
VERSION ?= 0.0.1

# IMAGE_BASE defines the image name for the project, without a version tag.
IMAGE_BASE ?= docker.greymatter.io/development/gm-operator

# Image urls to use for all building/pushing image targets.
# BUNDLE_IMG and CATALOG_IMG are OpenShift-specific.
IMG ?= $(IMAGE_BASE):$(VERSION)
BUNDLE_IMG ?= $(IMAGE_BASE)-bundle:$(VERSION)
CATALOG_IMG ?= $(IMAGE_BASE)-catalog:v$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'.

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt",year="2022" paths="./..."

manifests: controller-gen ## Generate CRD objects. These work back to Kubernetes 1.11.
	$(CONTROLLER_GEN) crd:trivialVersions=true,preserveUnknownFields=false paths="./..." output:crd:artifacts:config=config/base/crd/bases

pkgmanifests: kustomize manifests ## Generates a 'dev' bundle to be pushed directly to an OpenShift cluster.
	cd config/base/deployment && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/olm/manifests | operator-sdk generate packagemanifests --package gm-operator --version $(VERSION)

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: generate manifests fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

build: test ## Build operator binary.
	go build -o bin/operator main.go
	rm -rf bin/cue.mod/
	cp -r pkg/version/cue.mod/ bin/cue.mod

##@ Tools

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	GOBIN=$(shell pwd)/bin GOFLAGS=-mod=readonly go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
# Uses curl to run an install script because kustomize's go.mod does not yet support go install.
	if ! [[ -f $(KUSTOMIZE) ]]; then cd bin && curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh | bash;fi

##@ OpenShift Bundle

.PHONY: opm
OPM = ./bin/opm
opm: # Download opm locally if necessary. NOTE: This is hidden from the help target.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

.PHONY: bundle-run
bundle-run: ## Run the bundle image on the current OpenShift cluster. Uses oc under the hood.
	operator-sdk run bundle -n gm-operator --pull-secret-name gm-docker-secret $(BUNDLE_IMG)

##@ OpenShift Catalog

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	docker push $(CATALOG_IMG)
