VERSION=v1

build-dev:
	docker build -t  vahagnmian/example-webhook:${VERSION} --platform=linux/amd64 -f Dockerfile.dev .
	docker push vahagnmian/example-webhook:${VERSION}

build-prod:
	docker build -t  vahagnmian/example-webhook:${VERSION} --platform=linux/amd64 -f Dockerfile.prod .
	docker push vahagnmian/example-webhook:${VERSION}

cleanup:
	docker rmi $(docker images --filter=reference="*webhook:*" -q)