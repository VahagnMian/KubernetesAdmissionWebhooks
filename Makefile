VERSION=v1
REGISTRY=vahagnmian
REPO=webhook
VERSION=v1.0.0


build-dev:
	docker build -t  ${REGISTRY}/webhook:${VERSION} --platform=linux/amd64 -f Dockerfile.dev ./sourcecode
	docker push ${DOCKER_USERNAME}/webhook:${VERSION}

build-prod:
	docker build -t  ${DOCKER_USERNAME}/webhook:${VERSION} --platform=linux/amd64 -f Dockerfile.prod ./sourcecode
	docker push ${DOCKER_USERNAME}/webhook:${VERSION}

deploy:
	./refresh-certs.sh
	helm upgrade --install webhooks -n default -f ./webhooks/. \
		--set controller.image.registry="${REGISTRY}" \
		--set controller.image.repo="${REPO}" \
		--set controller.image.tag="${VERSION}"

cleanup:
	helm uninstall webhooks -n default
	docker rmi $(docker images --filter=reference="*webhook:*" -q)