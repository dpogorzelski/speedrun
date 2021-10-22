#!/bin/bash
set -e

openssl ecparam -name secp384r1 -genkey -noout -out portal.key
openssl req -new -key portal.key -out portal.csr -config <(
    cat <<-EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = SE
O = Speedrun
CN = portal

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
IP.1 = 34.74.21.99
EOF
)

openssl x509 -req -in portal.csr -CA ca.crt -CAkey ca.key -out portal.crt -days 365 -sha512 -CAcreateserial -extensions v3_req -extfile <(
    cat <<-EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = SE
O = Speedrun
CN = portal

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
IP.1 = 34.74.21.99
EOF
)
rm portal.csr
