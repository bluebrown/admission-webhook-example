#!/usr/bin/env bash

CERT_PATH="${1:-./certs}"
NAMESPACE="${2:-example}"
SERVICE="${3:-admission-webhook}"

mkdir -p "$CERT_PATH"

openssl req -x509 -newkey rsa:2048 -nodes -keyout "$CERT_PATH/ca.key" -out "$CERT_PATH/ca.crt" \
  -subj "/C=EU/CN=$SERVICE-ca.$NAMESPACE.svc" \
  -addext "subjectAltName = DNS:$SERVICE-ca.$NAMESPACE.svc"

openssl req -newkey rsa:2048 -nodes -keyout "$CERT_PATH/tls.key" -out "$CERT_PATH/server.csr" \
  -subj "/C=EU/CN=$SERVICE.$NAMESPACE.svc" \
  -addext "subjectAltName = DNS:$SERVICE.$NAMESPACE.svc"

openssl x509 -req -in "$CERT_PATH/server.csr" -CA "$CERT_PATH/ca.crt" \
  -CAkey "$CERT_PATH/ca.key" -CAcreateserial -out "$CERT_PATH/tls.crt" -days 365 -sha256 \
  -extfile <(printf "subjectAltName=DNS:%s.%s.svc" "$SERVICE" "$NAMESPACE")

rm -f "$CERT_PATH/server.csr" "$CERT_PATH/ca.srl"

chmod 644 "$CERT_PATH/tls.crt" "$CERT_PATH/tls.key"
