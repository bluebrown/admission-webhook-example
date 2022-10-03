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

.PHONY: build
build: ## Build the container image
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY) .

.PHONY: push
push: ## Push the container image to the image registry
	docker push $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)


##@ Deployment

.PHONY: install
install: gencert ## Deploy the webhook server and patch the pod and webhook config with a cert
	@./hack/deploy.sh $(certpath) $(namespace) $(service)

.PHONY: uninstall
uninstall: ## Remove the kubernetes resources
	kubectl delete -R -f config/ --ignore-not-found

.PHONY: gencert
gencert:  ## Generate a new certificate used for the webhooks
	./hack/gencert.sh $(certpath) $(namespace) $(service)

.PHONY: example
example: ## Deploy a sample configmap to see the webhooks in action
	kubectl apply -f config/example/

.PHONY: clean
clean: uninstall ## Clean up all the kubernetes resources, including example
	kubectl delete secret/admission-webhook --ignore-not-found
