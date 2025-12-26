# wpam - Werbot PAM Module

PAM (Pluggable Authentication Module) for Linux/Unix systems that provides two-factor authentication via Werbot service for SSH logins.

## Description

wpam integrates with the PAM authentication system and adds two-factor authentication (2FA) for SSH connections. The module supports:
- **TOTP** (Time-based One-Time Password) - 6-digit codes
- **U2F** (Universal 2nd Factor) - hardware security keys
- Offline mode for specified users
- Centralized access management via Werbot API

## Requirements

### For compilation:
- Go 1.25 or higher
- C compiler (gcc)
- PAM development headers

### For installation:
- Linux/Unix system with PAM support
- Access to Werbot API server

## Installation

### 1. Install dependencies

**Ubuntu/Debian:**
```bash
sudo apt install build-essential
sudo apt-get install pkg-config
sudo apt-get install libpam0g-dev
```

**CentOS/RHEL:**
```bash
sudo yum install gcc
sudo yum install pkgconfig
sudo yum install pam-devel
```

**FreeBSD:**
```bash
sudo pkg install go
sudo pkg install gcc
sudo pkg install pam-devel
```

**OpenBSD:**
```bash
doas pkg_add go gcc
# PAM is part of base system
```

**NetBSD:**
```bash
sudo pkgin install go gcc
sudo pkgin install pam-devel
```

### 2. Compilation

**For current platform:**
```bash
go build -buildmode=c-shared -o wpam.so
```

Or using Makefile:
```bash
make build
```

**Cross-compilation for specific platforms:**
```bash
# Linux amd64
make build-linux

# Linux arm64
make build-linux-arm64

# FreeBSD amd64
make build-freebsd

# OpenBSD amd64
make build-openbsd

# NetBSD amd64
make build-netbsd

# All platforms
make build-all
```

### 3. Install module

**Linux:**
```bash
sudo cp wpam.so /lib/security/pam_wpam.so
sudo chmod 644 /lib/security/pam_wpam.so
```

**FreeBSD/OpenBSD/NetBSD:**
```bash
sudo cp wpam.so /usr/lib/pam_wpam.so
sudo chmod 644 /usr/lib/pam_wpam.so
```

Or using Makefile (builds and installs for current platform):
```bash
make install
```

**Note:** PAM module directory varies by platform:
- Linux: `/lib/security/` or `/lib64/security/`
- FreeBSD/OpenBSD/NetBSD: `/usr/lib/`
- macOS: `/usr/lib/pam/` (not recommended for production)

### 4. Configure PAM

Edit `/etc/pam.d/sshd` and add the lines after `@include common-auth`:

```
auth required /lib/security/pam_wpam.so server_url=api.werbot.com service_id=YOUR_SERVICE_ID service_key=YOUR_SERVICE_KEY
account required /lib/security/pam_wpam.so server_url=api.werbot.com service_id=YOUR_SERVICE_ID service_key=YOUR_SERVICE_KEY
#@include common-auth
```

**Important:** 
- Comment out or remove the `@include common-auth` line if you want to use only Werbot for authentication
- `auth required` performs 2FA authentication (requires TOTP/U2F)
- `account required` performs account validation (checks access without TFA)

### 5. Configure SSH

Edit `/etc/ssh/sshd_config`:

```
ChallengeResponseAuthentication yes
UsePam yes
```

Restart SSH service:

**Linux (systemd):**
```bash
sudo systemctl restart sshd
```

**FreeBSD/OpenBSD/NetBSD:**
```bash
sudo service sshd restart
# or
sudo /etc/rc.d/sshd restart
```

## Configuration Parameters

All parameters are passed via PAM module arguments in `key=value` format:

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `server_url` | Yes | Werbot API server URL (without protocol) | `api.werbot.com` |
| `service_id` | Yes | Service ID in Werbot | `bef57548-dfdd-4aba-a545-50aa9e4f50db` |
| `service_key` | Yes | Service key in Werbot | `7ba94fcbac95b794fc6efc25e1c23f0d6b` |
| `offline_users` | No | Comma-separated list of users for offline access | `admin,backup` |
| `debug` | No | Enable debug logging (`true`/`false` or `1`/`0`) | `debug=true` |
| `insecure_skip_verify` | No | Disable SSL certificate verification (`true`/`false` or `1`/`0`) | `insecure_skip_verify=false` |

### Full configuration example:

```
auth required /lib/security/pam_wpam.so \
    server_url=api.werbot.com \
    service_id=bef57548-dfdd-4aba-a545-50aa9e4f50db \
    service_key=7ba94fcbac95b794fc6efc25e1c23f0d6b \
    offline_users=admin,backup \
    debug=false \
    insecure_skip_verify=false
```

## Usage

### Authentication Process

1. User connects via SSH
2. System prompts for **Werbot ID** (email or username in Werbot)
3. System prompts for **TOTP code** (6 digits) or empty string for U2F
4. Module sends request to Werbot API
5. On successful authentication, access is granted

### Session example:

```
$ ssh user@server
Enter your Werbot ID (email or username): user@example.com
Enter your totp code (submit empty for U2F): 123456
```

Or for U2F:
```
$ ssh user@server
Enter your Werbot ID (email or username): user@example.com
Enter your totp code (submit empty for U2F): [press Enter]
[Connect U2F key and press button]
```

## Logging

The module writes logs to `/var/log/wpam.log`. To view:

```bash
sudo tail -f /var/log/wpam.log
```

### Log levels:

- **INFO** - informational messages (offline access, successful authentication)
- **WARN** - warnings (failed authentication)
- **ERROR** - errors (connection issues, parsing problems)
- **DEBUG** - debug information (only when `debug=true`)

### Log file permissions:

```bash
sudo chmod 640 /var/log/wpam.log
sudo chown root:adm /var/log/wpam.log
```

## Security

### Recommendations:

1. **Disable debug in production**: `debug=false` (default)
2. **Do not use `insecure_skip_verify`** in production environment
3. **Limit offline access**: specify only necessary users in `offline_users`
4. **Protect service_key**: do not log or transmit in plain text
5. **Check module permissions**: `/lib/security/pam_wpam.so` should be accessible only to root

### Sensitive data:

The module automatically hides sensitive data in logs:
- `service_key` is replaced with `[REDACTED]`
- `totp_code` is replaced with `[REDACTED]`
- Tokens and keys in API responses are also hidden

## Offline Mode

If the Werbot server is unavailable, the module can grant access for specified users:

```
offline_users=admin,backup,emergency
```

**Warning:** Use offline mode only for critical accounts and only when necessary.

## Troubleshooting

### Issue: "PAM conversation error"

**Solution:** Ensure that `/etc/ssh/sshd_config` has:
```
ChallengeResponseAuthentication yes
UsePam yes
```
After changes, restart SSH: `sudo systemctl restart sshd`

### Issue: Module fails to load

**Solution:**
1. Check file permissions: `ls -l /lib/security/pam_wpam.so`
2. Check syntax in `/etc/pam.d/sshd`
3. Check logs: `sudo tail -f /var/log/wpam.log`

### Issue: Cannot connect to Werbot API

**Solution:**
1. Check server availability: `curl https://api.werbot.com`
2. Verify `server_url`, `service_id`, `service_key` are correct
3. Check firewall rules
4. Enable `debug=true` for detailed logging

### Issue: Authentication always fails

**Solution:**
1. Check logs: `sudo tail -f /var/log/wpam.log`
2. Ensure user is registered in Werbot
3. Verify TOTP code or U2F key is correct
4. Check service settings in Werbot

## Development

### Build:

**Current platform:**
```bash
make build
# or
go build -buildmode=c-shared -o wpam.so
```

**Cross-compilation:**
```bash
make build-linux      # Linux amd64
make build-linux-arm64 # Linux arm64
make build-freebsd    # FreeBSD amd64
make build-openbsd    # OpenBSD amd64
make build-netbsd     # NetBSD amd64
make build-all        # All platforms
```

**Help:**
```bash
make help
```

### Install (builds and installs):

```bash
make install
```

### Clean:

```bash
make clean
```

### Dependencies:

The project uses only Go standard library (`net/http`), no external dependencies required.

## License

[Specify project license]

## Support

[Specify support contacts]
