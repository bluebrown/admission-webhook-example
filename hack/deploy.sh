#!/usr/bin/env bash

# NOTE:
# by default the jobs are used so this script is not run
# it is here for reference only.

# use this, if you dont want to use the certgen jobs.
# it shows the required steps to make the webhooks work.
# the server must serve a certificate and the certificate
# must be present in the webhooks config.

set -euo pipefail

CERT_PATH="${1:-./certs}"
NAMESPACE="${2:-default}"
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
