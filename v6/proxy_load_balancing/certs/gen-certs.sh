#!/bin/sh
# Generate a CA + server cert for the proxy (idempotent).
# SANs cover the internal alias `proxy` and localhost so the same cert works
# both inside the compose network and from the host.
set -e

cd /certs

if [ -f server.crt ] && [ -f server.key ] && [ -f ca.crt ]; then
  echo "certs already present, skipping generation"
  exit 0
fi

apk add --no-cache openssl >/dev/null 2>&1 || true

echo "generating CA..."
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout ca.key -out ca.crt -days 3650 \
  -subj "/CN=Centrifugo Proxy LB Demo CA"

echo "generating server cert..."
openssl req -newkey rsa:2048 -nodes \
  -keyout server.key -out server.csr \
  -subj "/CN=proxy" \
  -addext "subjectAltName=DNS:proxy,DNS:localhost,IP:127.0.0.1"

openssl x509 -req -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 3650 -copy_extensions copyall

# HAProxy wants the cert and key concatenated in a single PEM file.
cat server.crt server.key > server.pem

rm -f server.csr ca.srl
chmod 644 *.crt *.key *.pem
echo "certs generated in /certs"
