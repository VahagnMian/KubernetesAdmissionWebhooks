VERSION=v1
REGISTRY=vahagnmian
REPO=webhook
VERSION=v1.0.0
LOCAL_IP=""
MUTATING_MANIFEST_FILE=./manifests/mutating-webhook-template.yaml
VALIDATING_MANIFEST_FILE=./manifests/validating-webhook-template.yaml


build-dev:
	docker build -t  ${REGISTRY}/webhook:${VERSION} --platform=linux/amd64 -f ./sourcecode/Dockerfile.dev ./sourcecode
	docker push ${REGISTRY}/webhook:${VERSION}

build-prod:
	docker build -t  ${REGISTRY}/webhook:${VERSION} --platform=linux/amd64 -f ./sourcecode/Dockerfile.prod ./sourcecode
	docker push ${REGISTRY}/webhook:${VERSION}


refresh-certs:
	./refresh-certs.sh

.PHONY: modify-manifest
modify-manifest:
ifeq ($(LOCAL_IP),"")
	#Modifing validating manifest file to send request to pod service
	$(eval TMP_FILE := $(shell mktemp))
	@gsed -i'' '/clientConfig:/,/caBundle:/{/service:/,/path:/d}' $(VALIDATING_MANIFEST_FILE)
	@gsed -i'' '/clientConfig:/,/caBundle:/{/url:/d}' $(VALIDATING_MANIFEST_FILE)
	@awk '/clientConfig:/{print;print "      service:\n        name: webhook\n        namespace: webhook\n        path: \"/validate\"";next}1' $(VALIDATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(VALIDATING_MANIFEST_FILE)

	#Modifing mutating manifest file to send request to pod service
	@gsed -i'' '/clientConfig:/,/caBundle:/{/service:/,/path:/d}' $(MUTATING_MANIFEST_FILE)
	@gsed -i'' '/clientConfig:/,/caBundle:/{/url:/d}' $(MUTATING_MANIFEST_FILE)
	@awk '/clientConfig:/{print;print "      service:\n        name: webhook\n        namespace: webhook\n        path: \"/mutate\"";next}1' $(MUTATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(MUTATING_MANIFEST_FILE)
else
	#Modifing validating manifest file to send request to development server
	$(eval TMP_FILE := $(shell mktemp))
	@awk '/service:/ {p=1} !p; /path:/ {p=0}' $(VALIDATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(VALIDATING_MANIFEST_FILE)
	@gsed -i'' '/service:/,+3d' $(VALIDATING_MANIFEST_FILE)
	@gsed -i'' '/clientConfig:/,/caBundle:/{/url:/d}' $(VALIDATING_MANIFEST_FILE)
	@awk '/clientConfig:/{print $$0 RS "      url: \"https://${LOCAL_IP}:8443/validate\"";next}1' $(VALIDATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(VALIDATING_MANIFEST_FILE)

	#Modifing mutating manifest file to send request to development server
	$(eval TMP_FILE := $(shell mktemp))
	@awk '/service:/ {p=1} !p; /path:/ {p=0}' $(MUTATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(MUTATING_MANIFEST_FILE)
	@gsed -i'' '/service:/,+3d' $(MUTATING_MANIFEST_FILE)
	@gsed -i'' '/clientConfig:/,/caBundle:/{/url:/d}' $(MUTATING_MANIFEST_FILE)
	@awk '/clientConfig:/{print $$0 RS "      url: \"https://${LOCAL_IP}:8443/mutate\"";next}1' $(MUTATING_MANIFEST_FILE) > $(TMP_FILE) && mv $(TMP_FILE) $(MUTATING_MANIFEST_FILE)
	@cd ./sourcecode && go run ./
endif

run:
	cd ./sourcecode
	go run ./

deploy: refresh-certs
	helm upgrade --install webhooks -n default -f ./webhooks/values.yaml  \
		--set controller.image.registry="${REGISTRY}" \
		--set controller.image.repo="${REPO}" \
		--set controller.image.tag="${VERSION}" ./webhooks/.
	make modify-manifest

cleanup:
	helm uninstall webhooks -n default
	rm -r ./webhooks/templates/mutating-webhook.yaml
	rm -r ./webhooks/templates/validating-webhook.yaml
	rm -r ./webhooks/templates/webhook-tls-secret.yaml