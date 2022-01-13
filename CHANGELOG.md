## 0.3.0-beta1 (2022-01-13)

### Breaking changes
Pretty much complete project pivot and no retro compatibility with previous versions. SSH support was dropped completely and an agent (Portal) was introduced.
You can refer to the README for updated installation and usage info.

## 0.2.2 (2021-08-01)

### Improvements
* Bump Go version to 1.17rc1 and replace `--filter` flag with `--target` and add a number of minor output tweaks ([#53](https://github.com/dpogorzelski/speedrun/issues/53))


## 0.2.1 (2021-07-06)

### Improvements

* Don't start the progress indicator if the log level is debug or lower (#52)


## 0.2.0 (2021-07-04)

### Features

* It's now possible to use the `--use-oslogin` flag with the `key authorize` and `run` command to use OS Login based authentication instead of the default method which adds the public key to the project metadata. In this mode the public key is added to the list of authorized keys of the user associated with the GOOGLE_APPLICATION_CREDENTIALS instead. ([#50](https://github.com/dpogorzelski/speedrun/issues/50))
* A progress indicator is now visible when using the `run` command ([#44](https://github.com/dpogorzelski/speedrun/issues/44))

### Improvements

* Confirm private key creation ([#45](https://github.com/dpogorzelski/speedrun/issues/45))
* Pass version information via ldflags ([#43](https://github.com/dpogorzelski/speedrun/issues/43))
* Suppress help on error ([#46](https://github.com/dpogorzelski/speedrun/issues/46))
* Wrap key type to b64 ([#51](https://github.com/dpogorzelski/speedrun/issues/51))

### Fixes

* Fix progress output ([#48](https://github.com/dpogorzelski/speedrun/issues/48))


## 0.1.1 (2021-03-28)

### Improvements

* Fetch home explicitly rather than expanding $HOME ([#41](https://github.com/dpogorzelski/speedrun/issues/41))
* Remove dead code

### Fixes

* Handle results pagination correctly ([#35](https://github.com/dpogorzelski/speedrun/issues/35))


## v0.1.0 (2021-03-21)

First usable release
