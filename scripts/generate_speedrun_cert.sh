#!/bin/bash
set -e

openssl ecparam -name secp384r1 -genkey -noout -out speedrun.key
openssl req -new -key speedrun.key -out speedrun.csr -config <(
    cat <<-EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = SE
O = Speedrun
CN = speedrun

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = clientAuth
EOF
)

openssl x509 -req -in speedrun.csr -CA ca.crt -CAkey ca.key -out speedrun.crt -days 365 -sha512 -CAcreateserial -extensions v3_req -extfile <(
    cat <<-EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = SE
O = Speedrun
CN = speedrun

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = clientAuth
EOF
)
rm speedrun.csr
