# Security Policy

## Supported versions

The latest minor release receives security fixes. Older versions: upgrade.

| Version | Supported |
|---|---|
| latest release | yes |
| anything older | no |

## Token handling

- `api_access_token` values are stored in the OS keyring (macOS Keychain, Linux
  Secret Service, Windows Credential Manager) under the service name `wootctl`,
  keyed by profile. Platform app tokens use the `<profile>/platform` key.
- On hosts without a keyring, tokens fall back to an AES-256-GCM encrypted file
  (`credentials.enc`, mode 0600) under the config dir. The key derives from
  `WOOTCTL_KEYRING_PASSWORD` via scrypt when set; otherwise from a host-bound seed,
  which is obfuscation, not a security boundary — set the password on shared hosts.
- Tokens never appear in `config.yaml`, command output, or `--dry-run` curls
  (redacted unless you pass `--show-token`).
- Cleartext `http://` base URLs are rejected for non-loopback hosts.
- `WOOTCTL_API_KEY` / `WOOTCTL_PLATFORM_TOKEN` env vars override the keyring for CI;
  treat your CI secret store accordingly.

## Reporting a vulnerability

Please use GitHub's private vulnerability reporting on this repository
(Security → Report a vulnerability). Do not open a public issue for anything
sensitive. Reports get a response within a week.
