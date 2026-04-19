CLI tool to sync HashiCorp Vault secrets into local `.env` files safely

---

## Installation

```bash
go install github.com/yourname/vaultpull@latest
```

Or download a prebuilt binary from the [Releases](https://github.com/yourname/vaultpull/releases) page.

---

## Usage

Authenticate with Vault and pull secrets into a local `.env` file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

vaultpull --path secret/data/myapp --output .env
```

This will fetch all key-value pairs from the specified Vault path and write them to `.env` in the format:

```
KEY=value
ANOTHER_KEY=another_value
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--path` | Vault secret path | _(required)_ |
| `--output` | Output file path | `.env` |
| `--addr` | Vault server address | `$VAULT_ADDR` |
| `--token` | Vault token | `$VAULT_TOKEN` |
| `--overwrite` | Overwrite existing file | `false` |

---

## Example

```bash
vaultpull --path secret/data/production --output .env.production --overwrite
```

---

## License

[MIT](LICENSE) © yourname