<p align="center">
  <a rel="nofollow">
    <img src="docs/logo.png?raw=true" width="200" style="max-width:100%;">
  </a>
</p>


# Speedrun
[![license](https://img.shields.io/badge/license-MPL2-blue.svg)](https://github.com/dpogorzelski/speedrun/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dpogorzelski/speedrun)](https://goreportcard.com/report/github.com/dpogorzelski/speedrun)
[![Go](https://github.com/dpogorzelski/speedrun/actions/workflows/go.yml/badge.svg)](https://github.com/dpogorzelski/speedrun/actions/workflows/go.yml)

Speedrun is an action execution framework that works at scale.

It allows you to run any action/command across any number of servers, projects and cloud vendors with ease and fast (currently GCP only but AWS and Azure will be supported as well), example:

```bash
speedrun run systemctl stop nginx
```

 to stop nginx across 3k machines.

No hassles with setting up and maintaining a server with agents as speedrun has none. Single self-contained binary. Speedrun leverages SSH as transport with tight cloud vendor integration to take the burden of mundane things like key generation and propagation/revocation away from the user.

Server targeting (`--target`) is made as intuitive as possible, currently based on [gcloud topic filters](https://cloud.google.com/sdk/gcloud/reference/topic/filters) but in the future will be replaced by a generic selection mechanism to provide a seamless experience across different cloud vendors.


Features:

* native cloud integration (currently Google Cloud only, AWS and Azure coming up!)
* stateless and agentless
* no complex configuration required
* single self-contained binary
* can run against any number of servers and projects seamlessly
* a plugin system is in the plan to allow anyone to integrate execution modules that will wrap complex functionality instead of running raw shell commands


## Installation

#### Homebrew (MacOS, Linux)
```bash
brew install dpogorzelski/tap/speedrun
```

#### Manual (MacOS, Linux, Windows)
Download the precompiled binary from here: [Releases](https://github.com/dpogorzelski/speedrun/releases)
## Usage

#### Quickstart
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount.json
speedrun init
speedrun key new
speedrun key authorize
speedrun run whoami
```

#### Portals

`sysctl -w net.core.rmem_max=2500000`

## Examples

Use 1000 concurrent SSH workers

```bash
speedrun run "uname -r" --concurrency 1000
```

Stop Nginx on VMs matching the target selector

```bash
speedrun run sudo systemctl stop nginx --target "labels.env=staging AND labels.app=foobar"
```

Ignore SSH fingerprint mismatch and connect via private IP addresses

```bash
speedrun run "ls -la" --target "labels.env != prod" --ignore-fingerprint --concurrency 1000 --use-private-ip
```

Use a different config file

```bash
speedrun run whoami -c /path/to/config.toml
```

## Configuration

Using certain flags repeteadly can be annoying, it's possible to persist their behavior via config file. Default config file is located at `~/.speedrun/config.toml` and can be re-initialized to it's default form via `speedrun init`.

#### Run it as nomad job


#### Run it as a systemd unit


#### Use self signed certificates

CA:
openssl ecparam -name secp384r1 -genkey -noout -out ca.key
openssl req -new -x509 -key ca.key -out ca.crt -days 1825 -subj "/C=SE/ST=/L=/O=Speedrun/OU=/CN=/"  -config <(
cat <<-EOF
[req]
default_bits = 2048
default_md = sha512
distinguished_name = dn
[ dn ]
[alt_names]
URI.1 = speedrun://local
[v3_req]
subjectKeyIdentifier=hash
basicConstraints=critical,CA:TRUE
keyUsage=critical,keyCertSign,cRLSign
subjectAltName = @alt_names
EOF
) -extensions 'v3_req'

portal cert:
openssl ecparam -name secp384r1 -genkey -noout -out portal.key
openssl req -new -key portal.key -out portal.csr -days 365 -subj "/C=SE/ST=/L=/O=Portal/OU=/CN=/" -extensions SAN -config <(cat /etc/ssl/openssl.cnf <(printf "[SAN]\nsubjectAltName='IP:34.74.21.99'"))
openssl x509 -req -in portal.csr -CA ca.crt -CAkey ca.key -out portal.crt -days 365 -sha256 -CAcreateserial

speedrun cert:
openssl ecparam -name secp384r1 -genkey -noout -out speedrun.key
openssl req -new -key speedrun.key -out speedrun.csr -days 365 -subj "/C=SE/ST=/L=/O=Speedrun/OU=/CN=/"
openssl x509 -req -in speedrun.csr -CA ca.crt -CAkey ca.key -out speedrun.crt -days 365 -sha256 -CAcreateserial

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Project Status

This project is in a very early stage so expect a lot of breaking changes in the nearest future

## Community & Support

Join the [#Speedrun](https://discord.gg/nkVvPnRvrJ) channel on Discord to chat and ask questions 😃

## License

[MPL-2.0](LICENSE)
