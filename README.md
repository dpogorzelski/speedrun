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
* extensible (plugin system is in the works)

## Installation

#### MacOS, Linux, Windows

Download the precompiled binaries from here: [Releases](https://github.com/dpogorzelski/speedrun/releases)

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

## Architecture

* picture first here

* speedrun client

* portal

* protocols

* service discovery

* language definition [expr/Language-Definition.md at master · antonmedv/expr · GitHub](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md)

## Examples

Stop Nginx on VMs that have a label `role` with value `nginx` and a label named `project` with value `someproject`

```bash
speedrun service stop nginx --target "labels.role == 'nginx' and labels.project == 'someproject'r"
```

Run arbitrary shell command on the target machines. Ignore Portal's certificate and connect via private IP address.

```bash
speedrun run "ls -la" --target "labels.env != 'prod'" --insecure --use-private-ip
```

Use a different config file

```bash
speedrun run whoami -c /path/to/config.toml
```

## Configuration

Instead of supplying certain flags repeatedly you can persist their behavior in the config file. Default config file is located at `~/.speedrun/config.toml` and default settings can be fetched via `speedrun init --print`.

#### Run portal as a systemd unit

#### Use self signed certificates during testing

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Project Status

This project is in a very early stage so expect a lot of breaking changes in the nearest future

## Community & Support

Join the [#Speedrun](https://discord.gg/nkVvPnRvrJ) channel on Discord to chat and ask questions 😃

## License

[MPL-2.0](LICENSE)
