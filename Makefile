VERSION=v1
REGISTRY=vahagnmian
REPO=webhook
VERSION=v1.0.0


build-dev:
	docker build -t  ${REGISTRY}/webhook:${VERSION} --platform=linux/amd64 -f ./sourcecode/Dockerfile.dev ./sourcecode
	docker push ${REGISTRY}/webhook:${VERSION}

build-prod:
	docker build -t  ${REGISTRY}/webhook:${VERSION} --platform=linux/amd64 -f ./sourcecode/Dockerfile.prod ./sourcecode
	docker push ${REGISTRY}/webhook:${VERSION}


refresh-certs:
	./refresh-certs.sh

deploy:
	helm upgrade --install webhooks -n default -f ./webhooks/values.yaml  \
		--set controller.image.registry="${REGISTRY}" \
		--set controller.image.repo="${REPO}" \
		--set controller.image.tag="${VERSION}" ./webhooks/.

cleanup:
	helm uninstall webhooks -n default
	rm -r ./webhooks/templates/mutating-webhook.yaml
	rm -r ./webhooks/templates/validating-webhook.yaml
	rm -r ./webhooks/templates/webhook-tls-secret.yaml
	docker rmi $(docker images --filter=reference="*webhook*" -q)