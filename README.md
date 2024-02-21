# Kubernetes Admission Webhooks

### This is the very basic of how Admission Webhooks are working in Kubernetes

## Build

```bash
# For development, and debugging purposes
make build-dev REGISTRY=<my-registry> REPO=webhook VERSION=v1.0.1-dev

# For production, lightweigt image
make build-prod REGISTRY=<my-registry> REPO=webhook VERSION=v1.0.2
```

## Generate / Refresh certificates
As kube-api server is using CA of webhooks to validate the certificate used for TLS connection we should generate certs, which is automated here
#### Prerequisites
 - cfssl
```bash
make refresh-certs
```


## Deploy
This command will deploy helm chart which contains everything that you need to run for controller
```bash
 make deploy REGISTRY=<registry> REPO=webhook VERSION=v1.0.2
```