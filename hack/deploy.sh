#!/usr/bin/env bash

set -euo pipefail

CERT_PATH="${1:-./certs}"
NAMESPACE="${2:-sandbox}"
SERVICE="${3:-admission-webhooks}"

kubectl create secret tls "$SERVICE" \
  --cert="$CERT_PATH/tls.crt" \
  --key="$CERT_PATH/tls.key" \
  --namespace="$NAMESPACE" \
  --dry-run=client -o yaml --save-config |
  kubectl apply -f -

kubectl apply -f config/webhooks/

kubectl patch MutatingWebhookConfiguration/example-annotator --type json \
  --patch '[{"op": "add", "path": "/webhooks/0/clientConfig/caBundle", "value": "'"$(base64 -w 0 <"$CERT_PATH/ca.crt")"'"}]'
