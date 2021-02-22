# Speedrun

Speedrun is a cloud first command execution utility. It's goal is to provide a simple way to perform command execution on remote servers with almost no setup.


## Installation

Download the precompiled binary from here:

```bash
curl https://github.com/dawidpogorzelski/speedrun/releases/tag/0.1.0
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

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
