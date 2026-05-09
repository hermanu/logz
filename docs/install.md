# Installation

`logz` ships as a single static binary with no runtime dependencies. Pick the
channel that suits your environment.

## Homebrew (macOS, Linux)

```bash
brew install hermanu/tap/logz
```

This pulls from [hermanu/homebrew-tap][tap]. To upgrade later:

```bash
brew upgrade logz
```

[tap]: https://github.com/hermanu/homebrew-tap

## Scoop (Windows)

```powershell
scoop bucket add hermanu https://github.com/hermanu/scoop-bucket
scoop install logz
```

## Go

If you have Go 1.26+ installed:

```bash
go install github.com/hermanu/logz@latest
```

The binary lands in `$GOBIN` (or `$GOPATH/bin`).

## Pre-built binaries

Every release attaches archives for Linux, macOS, and Windows on amd64/arm64
to the [GitHub Releases page][releases]. Download, extract, and put `logz`
somewhere on your `PATH`.

```bash
# Linux x86_64 example
curl -L https://github.com/hermanu/logz/releases/latest/download/logz_Linux_x86_64.tar.gz \
  | tar -xz -C /usr/local/bin logz
```

[releases]: https://github.com/hermanu/logz/releases

## Linux packages (.deb / .rpm / .apk)

Each release also ships native packages:

```bash
# Debian / Ubuntu
curl -LO https://github.com/hermanu/logz/releases/latest/download/logz_<version>_linux_amd64.deb
sudo dpkg -i logz_<version>_linux_amd64.deb

# RHEL / Fedora / CentOS
sudo rpm -i https://github.com/hermanu/logz/releases/latest/download/logz_<version>_linux_amd64.rpm

# Alpine
apk add --allow-untrusted logz_<version>_linux_amd64.apk
```

## Docker

```bash
docker run --rm -i ghcr.io/hermanu/logz:latest filter --level error < app.log
```

Multi-arch images (`linux/amd64`, `linux/arm64`) are tagged on each release
and as `:latest`.

## From source

```bash
git clone https://github.com/hermanu/logz
cd logz
make build
./logz --version
```

Requires Go 1.26+.

## Verifying releases

Each release publishes a `checksums.txt` file. Verify a download with:

```bash
sha256sum -c checksums.txt --ignore-missing
```

A Software Bill of Materials (SBOM, SPDX format) is attached to each archive
under `<archive-name>.sbom.json`.
