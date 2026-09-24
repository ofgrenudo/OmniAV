#!/bin/sh
# Runs on container start (nginx image's /docker-entrypoint.d). If no certificate is present in
# /etc/nginx/certs, generate a 10-year self-signed one. A real certificate mounted there is used as-is.
set -eu

cert="/etc/nginx/certs/${TLS_CERT_FILE}"
key="/etc/nginx/certs/${TLS_KEY_FILE}"

if [ -s "$cert" ] && [ -s "$key" ]; then
    echo "generate-cert: using existing certificate $cert"
    exit 0
fi

echo "generate-cert: no certificate found, generating self-signed certificate (10 years)"
mkdir -p /etc/nginx/certs
openssl req -x509 -nodes -newkey rsa:2048 -days 3650 \
    -keyout "$key" -out "$cert" \
    -subj "/CN=${TLS_SERVER_NAME}/O=OmniAV" \
    -addext "subjectAltName=DNS:${TLS_SERVER_NAME},DNS:localhost,IP:127.0.0.1"
chmod 600 "$key"
