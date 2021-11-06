<p align="center">
  <a rel="nofollow">
    <img src="assets/logo.png?raw=true" width="200" style="max-width:100%;">
  </a>
</p>


# Speedrun
[![license](https://img.shields.io/badge/license-MPL2-blue.svg)](https://github.com/dpogorzelski/speedrun/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dpogorzelski/speedrun)](https://goreportcard.com/report/github.com/dpogorzelski/speedrun)
[![Go](https://github.com/dpogorzelski/speedrun/actions/workflows/go.yml/badge.svg)](https://github.com/dpogorzelski/speedrun/actions/workflows/go.yml)

Speedrun helps you control your compute fleet with minimal effort.

Example (stop nginx across 3k machines):

```bash
speedrun service stop nginx
```
Features:

* stateless
* serverless
* idempotent
* no complex configuration required
* server discovery via native cloud integration (currently Google Cloud only, AWS and Azure coming up!)

## Installation

#### Homebrew (MacOS, Linux)
```bash
brew install dpogorzelski/tap/speedrun
```

#### Manual (MacOS, Linux, Windows)
Download the precompiled binary from here: [Releases](https://github.com/dpogorzelski/speedrun/releases)
## Usage

#### Quickstart
On a server:

`sudo ./portal --insecure`

On your machine:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount.json
speedrun init
speedrun run whoami --insecure
```

## Examples

Use 1000 concurrent SSH workers

```bash
speedrun run "uname -r" --concurrency 1000
```

Stop Nginx on VMs matching the target selector

```bash
speedrun service stop nginx --target "labels.env=staging AND labels.app=foobar"
```

Ignore Portal's certificate and connect via private IP addresses

```bash
speedrun run "ls -la" --target "labels.env != prod" --insecure --use-private-ip
```

Use a different config file

```bash
speedrun run whoami -c /path/to/config.toml
```

## Configuration

Using certain flags repeteadly can be annoying, it's possible to persist their behavior via config file. Default config file is located at `~/.speedrun/config.toml` and can be re-initialized to it's default form via `speedrun init`.

#### Run it as a systemd unit


#### Use self signed certificates

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Project Status

This project is in a very early stage so expect a lot of breaking changes in the nearest future

## Community & Support

Join the [#Speedrun](https://discord.gg/nkVvPnRvrJ) channel on Discord to chat and ask questions 😃

## License

[MPL-2.0](LICENSE)
