#!/bin/bash

cfssl gencert -initca ./tls/ca-csr.json | cfssljson -bare /tmp/ca

cfssl gencert \
  -ca=/tmp/ca.pem \
  -ca-key=/tmp/ca-key.pem \
  -config=./tls/ca-config.json \
  -hostname="example-webhook,example-webhook.webhook.svc.cluster.local,example-webhook.webhook.svc,localhost,127.0.0.1,192.168.11.80" \
  -profile=default \
  ./tls/ca-csr.json | cfssljson -bare /tmp/example-webhook

cat /tmp/example-webhook.pem > ./tls/local-dev-certs/tls.crt
cat /tmp/example-webhook-key.pem > ./tls/local-dev-certs/tls.key

cat <<EOF > ./manifests/webhook-tls-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-webhook-tls
type: Opaque
data:
  tls.crt: $(cat /tmp/example-webhook.pem | base64 | tr -d '\n')
  tls.key: $(cat /tmp/example-webhook-key.pem | base64 | tr -d '\n')
EOF

ca_pem_b64="$(openssl base64 -A <"/tmp/ca.pem")"

sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"./manifests/mutating-webhook-template.yaml" \
    > "./manifests/mutating-webhook.yaml"

sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"./manifests/validating-webhook-template.yaml" \
    > "./manifests/validating-webhook.yaml"

export KUBECONFIG=/Users/vahagn/.kube/config31

kubectl apply -f ./manifests/webhook-tls-secret.yaml
kubectl delete -f ./manifests/mutating-webhook.yaml
kubectl apply -f ./manifests/mutating-webhook.yaml

kubectl delete -f ./manifests/validating-webhook.yaml
kubectl apply -f ./manifests/validating-webhook.yaml