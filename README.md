<p align="center">
  ![logo](docs/logo2.png)
</p>
# Speedrun

Speedrun is a command execution utility. It's goal is to provide a simple way to perform command execution on a large amount of servers.

Features:

* native cloud integration (currently Google Cloud only, AWS and Azure coming up!)
* incredibly fast
* stateless and agentless
* no complex configuration required


## Installation

Download the precompiled binary from here:

```bash
curl https://github.com/dpogorzelski/speedrun/releases/tag/0.1.0
mv speedrun /usr/local/bin
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

## License

[MIT](https://choosealicense.com/licenses/mit/)
