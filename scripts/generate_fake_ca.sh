#!/bin/bash
set -e

openssl ecparam -name secp384r1 -genkey -noout -out ca.key
openssl req -new -x509 -key ca.key -out ca.crt -days 365 -config <(
    cat <<-EOF
[req]
distinguished_name = req_distinguished_name
default_md = sha512
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = SE
O = Speedrun
CN = Fake CA

[v3_req]
subjectKeyIdentifier=hash
basicConstraints=critical,CA:TRUE
keyUsage=critical,keyCertSign,cRLSign
EOF
) -extensions v3_req
