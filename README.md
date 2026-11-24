# berglas9

TUI for [Berglas](https://github.com/GoogleCloudPlatform/berglas) secrets in a Cloud Storage bucket. List names, reveal a value, copy it, and create or update secrets. Values stay masked until you reveal them.

## Quick Start

Sign in with the gcloud CLI and select a project. berglas9 uses that account and project. It does not use Application Default Credentials.

```bash
gcloud auth login
gcloud config set project my-project
```

Build and run:

```bash
av make
./berglas9 my-bucket --key projects/my-p/locations/global/keyRings/my-kr/cryptoKeys/my-k
```

Both arguments are required. `<bucket>` is the Cloud Storage bucket. `--key` is the Cloud KMS key used when creating or updating a secret.

## Requirements

- **Go** 1.24+
- **gcloud CLI**, signed in with `gcloud auth login`, and an active project (`gcloud config set project`)
- Access to the bucket, and Cloud KMS CryptoKey Encrypter/Decrypter on the key, for that account
- A clipboard tool on Linux (`xclip`, `xsel`, or `wl-copy`). macOS uses `pbcopy`

## Usage

```text
berglas9 <bucket> --key=<kms-key>
berglas9 -h | --help
berglas9 -v | --version
```

### Keys

| Key | Action |
|-----|--------|
| **↑ / ↓** | Select a secret |
| **m** | Mask or unmask the selected value |
| **Ctrl+C** | Copy the selected secret on Linux and other systems. Does not quit |
| **⌘C** | Copy the selected secret on macOS. Ctrl+C quits |
| **n** | New secret: name, then value |
| **e** | Edit the selected secret |
| **r** | Refresh the list |
| **h** | Help |
| **q** | Quit |

Inside a create or edit dialog, the copy key cancels and does not copy. That key is **⌘C** on macOS and **Ctrl+C** elsewhere.

Secret values render as `••••••••` until revealed. The mask does not reflect the length of the value. The status line never prints a secret.

## Development

```bash
av tests
av make
```

The binary is `./berglas9`. `av release` builds the archives used by the Homebrew formula.
