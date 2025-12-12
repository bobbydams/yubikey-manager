# YubiKey GPG Manager

[![CI](https://github.com/bobbydams/yubikey-manager/actions/workflows/ci.yml/badge.svg)](https://github.com/bobbydams/yubikey-manager/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/bobbydams/yubikey-manager/graph/badge.svg?token=CWCHT0ITHM)](https://codecov.io/github/bobbydams/yubikey-manager)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobbydams/yubikey-manager)](https://goreportcard.com/report/github.com/bobbydams/yubikey-manager)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/bobbydams/yubikey-manager.svg)](https://github.com/bobbydams/yubikey-manager/releases)

A robust Go CLI tool for managing GPG signing subkeys across multiple YubiKeys. This tool provides a safe and convenient way to set up, manage, and maintain GPG keys on YubiKey hardware security modules.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**IMPORTANT DISCLAIMER**: This software is provided "AS IS" without warranty of any kind. The authors assume no liability for any loss or corruption of GPG keys, damage to YubiKey devices, or any other consequences resulting from the use of this software. **Use at your own risk.** Always maintain secure backups of your master GPG keys and test operations in a safe environment before using on production keys or devices.

## Features

- **Setup New YubiKeys**: Add signing subkeys to new YubiKeys (interactive or semi-automated)
- **Revoke Compromised Keys**: Revoke subkeys when YubiKeys are lost or compromised
- **Extend Expiration**: Extend expiration dates on keys and subkeys
- **Key Management**: Clean up old/expired keys from your keyring
- **Backup Management**: Automatic backups before making changes
- **Status & Verification**: Check key and YubiKey status, verify setup

## Installation

### Prerequisites

- Go 1.21 or later
- GPG 2.2+ installed and configured
- YubiKey with OpenPGP support

### Build from Source

```bash
git clone https://github.com/bobbydams/yubikey-manager.git
cd yubikey-manager
go build -o ykgpg ./cmd/ykgpg
```

### Install

```bash
go install ./cmd/ykgpg
```

## Configuration

### Interactive Configuration Setup

The easiest way to set up configuration is using the interactive command:

```bash
ykgpg config init
```

This will prompt you for all required configuration values and create the config file at `~/.config/ykgpg/config.yaml`.

### Manual Configuration File

Alternatively, you can manually create a configuration file at `~/.config/ykgpg/config.yaml`:

```yaml
primary_key_id: "YOUR_KEY_ID_HERE"
primary_key_fingerprint: "YOUR_FULL_FINGERPRINT_HERE"
user_name: "Your Name"
user_email: "your.email@example.com"
keyserver: "hkps://keys.openpgp.org"
backup_dir: "~/.gnupg/backups"
no_color: false # Set to true to disable colored output
```

You can copy `config.example.yaml` as a starting point.

### Disable Colors

You can disable colored output in several ways:

- **CLI flag**: `ykgpg --no-color <command>`
- **Config file**: Set `no_color: true` in your config file
- **Environment variable**: `export YKGPG_NO_COLOR=true`

### View Current Configuration

To see your current configuration values from all sources:

```bash
ykgpg config show
```

This displays the effective configuration values, showing which sources (config file, environment variables, CLI flags) are being used.

### Environment Variables

All configuration values can be overridden with environment variables using the `YKGPG_` prefix:

```bash
export YKGPG_PRIMARY_KEY_ID="YOUR_KEY_ID_HERE"
export YKGPG_MASTER_KEY_PATH="/path/to/master/key.asc"
```

### CLI Flags

Configuration can also be set via CLI flags:

```bash
ykgpg status --key-id "YOUR_KEY_ID_HERE" --name "Your Name"
```

**Priority Order**: CLI flags > Environment variables > Config file > Defaults

## Usage

### Show Status

```bash
ykgpg status
```

Displays information about your primary key and connected YubiKey.

### Setup New YubiKey

**Interactive mode** (recommended for first-time setup):

```bash
ykgpg setup
```

**Semi-automated mode** (faster, but still requires some interaction):

```bash
ykgpg setup-batch
```

Both commands will:

1. Create a backup of your current keyring
2. Prompt for your master key backup
3. Generate a new signing subkey
4. Move the subkey to your YubiKey
5. Optionally remove the master key from your local machine
6. Optionally upload the updated key to a keyserver

### Revoke a Subkey

If a YubiKey is lost or compromised:

```bash
ykgpg revoke
```

This will:

1. Show all current signing subkeys
2. Prompt you to select which key to revoke
3. Create a backup
4. Import your master key
5. Guide you through the revocation process
6. Upload the revocation to the keyserver

### Extend Key Expiration

```bash
ykgpg extend
```

Extends the expiration date on your primary key and all subkeys.

### Clean Up Old Keys

```bash
ykgpg cleanup
```

Helps identify and remove old or expired keys from your keyring.

### Set YubiKey Metadata

```bash
ykgpg set-metadata
```

Sets the cardholder name and URL on your YubiKey for easier identification.

### Export Public Key

```bash
ykgpg export
ykgpg export --output /path/to/key.asc
```

Exports your public key for sharing or uploading to keyservers.

### Verify Setup

```bash
ykgpg verify
```

Checks that your GPG and YubiKey setup is correct, including:

- Primary key exists
- Master key is offline
- YubiKey is detected
- Git signing configuration
- GPG signing works

## Commands

| Command        | Description                                            |
| -------------- | ------------------------------------------------------ |
| `status`       | Show current key and YubiKey status                    |
| `setup`        | Add a signing subkey to a new YubiKey (interactive)    |
| `setup-batch`  | Add a signing subkey to a new YubiKey (semi-automated) |
| `revoke`       | Revoke a subkey (for lost/compromised YubiKeys)        |
| `extend`       | Extend expiration dates on keys                        |
| `cleanup`      | Remove old/expired keys from keyring                   |
| `set-metadata` | Set cardholder name and URL on YubiKey                 |
| `export`       | Export public key to file                              |
| `verify`       | Verify GPG and YubiKey setup                           |
| `config init`  | Interactively generate configuration file              |
| `config show`  | Show current configuration values                      |

## Development

### Running Tests

**Unit tests** (with mocked dependencies):

```bash
go test ./...
```

**Integration tests** (requires GPG and optionally a YubiKey):

```bash
go test -tags=integration ./...
```

### Project Structure

```
yubikey-manager/
├── cmd/ykgpg/          # CLI entry point
├── internal/
│   ├── cli/            # CLI commands
│   ├── gpg/            # GPG service and parsers
│   ├── yubikey/        # YubiKey service
│   ├── backup/         # Backup service
│   ├── config/         # Configuration management
│   └── executor/       # Command execution abstraction
├── pkg/ui/             # UI helpers (output, prompts)
└── testdata/           # Test fixtures
```

## Security Considerations

- **Master Key Safety**: The master key should never be stored on your regular machine. Always use a secure backup (USB drive, encrypted storage).
- **Backups**: The tool automatically creates backups before making changes. Keep these backups secure.
- **Key Revocation**: Revoked keys cannot be un-revoked. Be certain before revoking a key.
- **YubiKey PINs**: Keep your YubiKey PINs secure and use strong PINs.

## Contributing

Contributions are welcome! Please open an issue or pull request.

## Acknowledgments

This tool was converted from a Bash script to provide better testability, maintainability, and cross-platform support.
