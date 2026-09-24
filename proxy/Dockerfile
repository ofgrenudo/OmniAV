FROM nginx:1.27-alpine

RUN apk add --no-cache openssl

# Defaults produce/expect tls.crt + tls.key in /etc/nginx/certs; override to match a real cert's
# filenames (e.g. Let's Encrypt's fullchain.pem / privkey.pem).
ENV TLS_CERT_FILE=tls.crt \
    TLS_KEY_FILE=tls.key \
    TLS_SERVER_NAME=localhost

COPY generate-cert.sh /docker-entrypoint.d/10-generate-cert.sh
COPY templates/ /etc/nginx/templates/

EXPOSE 443
