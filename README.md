<p align="center">
  <a rel="nofollow">
    <img src="docs/logo.png?raw=true" width="200" style="max-width:100%;">
  </a>
</p>


# Speedrun

Speedrun executes commands, at scale.

Features:

* native cloud integration (currently Google Cloud only, AWS and Azure coming up!)
* stateless and agentless
* no complex configuration required


## Installation

Download the precompiled binary from here:

```bash
curl https://github.com/dpogorzelski/speedrun/releases/tag/0.1.0
```

## Usage

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount.json
speedrun init
speedrun key new
speedrun key set
speedrun run whoami
```

## Examples

Use 1000 concurrent SSH workers

```bash
speedrun run "uname -r" --concurrency 1000
```

Stop Nginx on VMs matching the filter

```bash
speedrun run sudo systemctl stop nginx --filter "labels.env=staging AND labels.app=foobar"
```

Ignore SSH fingerprint mismatch and connect via private IP addresses

```bash
speedrun run "ls -la" --filter "labels.env != prod" --ignore-fingerprint --concurrency 1000 --use-private-ip
```

Use a different config file

```bash
speedrun run whoami -c /path/to/config.toml
```

Set public key on specific instances instead of project metadata (useful if instances are blocking project wide keys):

```bash
speedrun key set --filter "labels.env = dev"
```



## Configuration

Using certain flags repeteadly can be annoying, it's possible to persist their behavior via config file. Default config file is located at `~/.speedrun/config.toml` and can be re-initialized to it's default form via `speedrun init`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Project Status

This project is in a very early stage so expect a lot of breaking changes in the nearest future

## License

[MPL-2.0](LICENSE)