certpath=certs
namespace=sandbox
service=admission-webhooks

IMAGE_REGISTRY=index.docker.io
IMAGE_REPOSITORY=bluebrown/admission-webhooks-example

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


# this cool help text is from https://github.com/kubernetes-sigs/kubebuilder. Than you :)
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Build

.PHONY: all
all: build push ## Build and push

.PHONY: build
build: ## Build the container image
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY) .

.PHONY: push
push: ## Push the container image to the image registry
	docker push $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)


##@ Deployment

.PHONY: install
configure: ## Configure the namespace of the kustomization. i.e. make configure namespace=foo
	@cd config/deploy/ && kustomize edit set namespace $(namespace)

.PHONY: install
install: configure ## Deploy the webhook server and patch the pod and webhook config with a cert
	@kustomize build config/deploy/ | kubectl apply -f -

.PHONY: uninstall
uninstall: ## Remove the kubernetes resources (does NOT remove the namespace)
	@@kustomize build config/deploy/ | kubectl delete --ignore-not-found -f -


##@ Development

.PHONY: gencert
gencert:  ## Generate a new certificate used for the webhooks. Useful for local development
	./hack/gencert.sh $(certpath) $(namespace) $(service)

.PHONY: example
example-apply: ## Deploy a sample configmap to see the webhooks in action
	kubectl apply -f hack/example-cm.yaml

.PHONY: clean
example-delete: ## Clean up the example example
	kubectl delete --ignore-not-found -f hack/example-cm.yaml
