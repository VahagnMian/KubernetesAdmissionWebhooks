# Kubernetes Admission Webhooks

## TL;DR

## Build

```bash
# For development, and debugging purposes
make build-dev REGISTRY=<my-registry> REPO=webhook VERSION=v1.0.1-dev

# For production, lightweigt image
make build-prod REGISTRY=<my-registry> REPO=webhook VERSION=v1.0.4
```

## Deploy
This command will deploy helm chart which contains everything that you need to run for controller

This will automatically create, refresh, and replace certificates, and will also replace the template files to helm chart
```bash
kubectl create ns webhooks
make deploy REGISTRY=<registry> NAMESPACE=webhooks REPO=webhook VERSION=v1.0.4
```

## Development
For development purposes you can deploy this way, specifying the IP address your webhook should send requests to

This will automatically modify your manifests
```bash
make deploy LOCAL_IP="x.x.x.x"
```

## Cleanup
This will delete every generated file, and undeploy helm chart from k8s
```bash
make cleanup
```