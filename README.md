# bx

[![CI](https://github.com/maeshinshin/bx/actions/workflows/test.yml/badge.svg)](https://github.com/maeshinshin/bx/actions/workflows/test.yml)
[![Latest Release](https://img.shields.io/github/v/release/maeshinshin/bx)](https://github.com/maeshinshin/bx/releases/latest)
[![codecov](https://codecov.io/gh/maeshinshin/bx/graph/badge.svg)](https://app.codecov.io/gh/maeshinshin/bx)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`bx` is a small command-line tool for Base64 encoding and decoding, with first-class support for the `data` and `stringData` fields of Kubernetes Secret YAML manifests. It is a single static binary with no external dependencies and works without `kubectl`.

## Contents

- [Why bx?](#why-bx)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [Usage](#usage)
  - [encode](#encode)
  - [decode](#decode)
  - [decode Kubernetes Secret YAML](#decode-kubernetes-secret-yaml)
- [Behavior notes](#behavior-notes)
- [Commands and flags](#commands-and-flags)
- [Shell completion](#shell-completion)
- [Limitations](#limitations)
- [License](#license)

## Why bx?

`bx d` replaces `base64 -d` for quick decoding:

```sh
$ echo "aGVsbG8=" | bx d
hello
```

For Kubernetes Secrets, the alternative is a long `yq` query:

```sh
$ kubectl get secret my-secret -o yaml \
    | yq '.data | to_entries | .[] | "\(.key): \(.value | @base64d)"'
```

With `bx` it becomes the same one-liner as plain Base64 decoding:

```sh
$ kubectl get secret my-secret -o yaml | bx d -k
```

`bx` reads from arguments, files, or stdin, and turns the slow dance between `yq`, `jq`, and `base64 -d` into a single command. Use `bx e` / `bx d` as shortcuts for `bx encode` / `bx decode`.

## Installation

### Homebrew

```sh
brew install maeshinshin/tap/bx
```

### Download binary

Download the archive for your platform from the [Releases](https://github.com/maeshinshin/bx/releases) page, extract it, and place `bx` in a directory on your `PATH`.

The `releases/latest/download/` URLs always point to the latest published version, so you can paste these as-is:

```sh
# Linux amd64
curl -L -o bx.tar.gz https://github.com/maeshinshin/bx/releases/latest/download/bx_linux_amd64.tar.gz
tar -xzf bx.tar.gz
sudo install bx /usr/local/bin/

# macOS (Apple Silicon)
curl -L -o bx.tar.gz https://github.com/maeshinshin/bx/releases/latest/download/bx_darwin_arm64.tar.gz
tar -xzf bx.tar.gz
sudo install bx /usr/local/bin/
```

To pin a specific version, replace `latest` with the version tag (without the leading `v`):

```sh
VERSION=1.1.0
curl -L -o bx.tar.gz "https://github.com/maeshinshin/bx/releases/download/v${VERSION}/bx_${VERSION}_linux_amd64.tar.gz"
tar -xzf bx.tar.gz
sudo install bx /usr/local/bin/
```

Available targets are `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`, `windows_amd64`, and `windows_arm64` (Windows archives use `.zip` instead of `.tar.gz`).

On Windows, download `bx_<version>_windows_amd64.zip` (or `windows_arm64.zip`) and extract `bx.exe`.

### Build from source

Requires Go 1.23 or later.

```sh
git clone https://github.com/maeshinshin/bx.git
cd bx
go build -trimpath -ldflags "-s -w" -o bx .
mv bx /usr/local/bin/   # or any directory on your PATH
```

Or, if you only need the binary in your Go cache and have `$GOPATH/bin` on your PATH:

```sh
go install github.com/maeshinshin/bx@latest
```

Verify the installation:

```sh
bx --help
```

## Quickstart

```sh
# Encode a string
$ bx encode "hello"
aGVsbG8=

# Decode a Base64 string
$ bx decode "aGVsbG8="
hello

# Decode every value in the data field of a Secret
$ bx decode -k "data:
  username: YWRtaW4=
  password: cEBzc3dvcmQ=
"
username: admin
password: p@ssword
```

## Usage

`bx` accepts input from one of three sources, in this order of precedence:

1. A positional argument
2. A file passed with `-f` / `--file`
3. stdin (when not a terminal)

### encode

Encode a string to Base64 and print it to stdout.

```sh
$ bx encode "hoge"
aG9nZQ==
```

From a file:

```sh
$ bx encode -f input.txt
```

From stdin:

```sh
$ echo "hoge" | bx encode
aG9nZQ==
```

### decode

Decode a Base64 string and print the result to stdout.

```sh
$ bx decode "aG9nZQ=="
hoge
```

From a file:

```sh
$ bx decode -f encoded.txt
```

From stdin:

```sh
$ echo "aG9nZQ==" | bx decode
hoge
```

### decode Kubernetes Secret YAML

Use `-k` / `--k8s` to extract and decode every value in the `data` and `stringData` fields of a Secret manifest. Values under `data:` are Base64-decoded; values under `stringData:` are emitted verbatim. The output is `key: value` pairs, one per line.

```sh
$ bx decode -k "data:
  username: YWRtaW4=
  password: cEBzc3dvcmQ=
"
username: admin
password: p@ssword
```

`stringData` works the same way:

```sh
$ bx decode -k "stringData:
  username: admin
  password: p@ssword
"
username: admin
password: p@ssword
```

From a manifest file:

```sh
$ bx decode -k -f secret.yaml
```

From stdin (typical pipeline with `kubectl`):

```sh
$ kubectl get secret my-secret -o yaml | bx decode -k
```

## Behavior notes

- Output is always followed by a single trailing newline.
- A value under `data:` that is not valid Base64 is emitted unchanged and no error is raised. This lets you inspect partially-correct manifests without a hard failure.
- Values under `stringData:` are emitted verbatim with no Base64 decoding (they are already plain text by definition).
- The `data:` and `stringData:` keys must be YAML mappings. If either is missing, a sequence, or a scalar, the field is silently skipped.
- Multiple YAML documents in a single input are all processed and their results are concatenated (data first, then stringData).
- `bx` does not consult a Kubernetes cluster. It operates purely on the YAML text you provide.

## Commands and flags

| Command  | Alias | Description                                                   |
| -------- | ----- | ------------------------------------------------------------- |
| `encode` | `e`   | Encode a string to Base64                                     |
| `decode` | `d`   | Decode a Base64 string, or a Kubernetes Secret YAML with `-k` |

| Flag           | Commands           | Description                                                                                              |
| -------------- | ------------------ | -------------------------------------------------------------------------------------------------------- |
| `-f`, `--file` | `encode`, `decode` | Read input from a file instead of an argument or stdin                                                   |
| `-k`, `--k8s`  | `decode`           | Treat input as a Kubernetes Secret YAML and decode every value in the `data` and `stringData` fields     |
| `-u`, `--url`  | `encode`, `decode` | Use the URL-safe Base64 alphabet (`-_`) instead of the standard alphabet (`+/`)                          |
| `-h`, `--help` | all                | Show help for a command                                                                                  |

Exit codes:

- `0` on success
- `1` on any error (bad input, missing file, invalid Base64, invalid YAML)

## Shell completion

`bx` ships with cobra's built-in completion command. Load it from your shell startup file so completions work only for your user (no `sudo` required).

For **bash**, add this to `~/.bashrc`:

```sh
source <(bx completion bash)
```

For **zsh**, add this to `~/.zshrc`:

```sh
source <(bx completion zsh)
```

For **fish**, add this to `~/.config/fish/config.fish`:

```sh
bx completion fish | source
```

For **PowerShell**, add this to your profile (`$PROFILE`):

```powershell
bx completion powershell | Out-String | Invoke-Expression
```

Reload the shell (or `source ~/.bashrc`) to activate completions.

## Limitations

- Only the top-level `data` and `stringData` keys are decoded. Nested mappings and lists inside those keys are skipped.
- The output format for `decode -k` is a simple `key: value` text and is not a structured format such as JSON or YAML.

## License

Apache License 2.0. See [LICENSE](LICENSE) for the full text.
