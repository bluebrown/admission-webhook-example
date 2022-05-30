certpath=certs
namespace=example
service=admission-webhook

IMAGE_REGISTRY=docker.io
IMAGE_REPOSITORY=bluebrown/admission-webhook-example


deploy:
	./scripts/deploy.sh $(certpath) $(namespace) $(service)

example:
	kubectl apply -f manifests/example/

gencert:
	./scripts/gencert.sh $(certpath) $(namespace) $(service)

build:
	docker build -t $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY) .

push:
	docker push $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)

clean:
	kubectl delete -R -f manifests/ --ignore-not-found
	kubectl delete secret/admission-webhook --ignore-not-found
