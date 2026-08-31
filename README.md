# vm-template

A lightweight, zero-dependency, static Go utility for preparing and sanitizing Linux Virtual Machines (Ubuntu/Debian) before converting them into VMware vSphere Golden Templates.

## Features

- **Static Binary**: Built with `CGO_ENABLED=0` for zero runtime glibc/NSS dependencies.
- **State Inspection**: `inspect` subcommand checks machine-id size, SSH host keys, VMware Tools status, and Netplan configurations.
- **Automated Sanitization**:
  - Stops logging daemons (`systemd-journald`, `rsyslog`)
  - Resets `cloud-init` state and logs (if present)
  - Truncates `/etc/machine-id` (0 bytes) and fixes `/var/lib/dbus/machine-id` symlink
  - Purges existing SSH host keys (`/etc/ssh/ssh_host_*`)
  - Configures first-boot SSH host key auto-regeneration systemd service
  - Cleans and re-configures Netplan to use universal DHCP without MAC locking
  - Purges APT caches, `/var/log/*`, `/tmp/*`, and user shell histories (`~/.bash_history`)
  - Optional automatic shutdown (`--poweroff`)
- **Safe Dry-Run**: Test execution without modifying files using `--dry-run`.

## Building

```bash
# Build static 64-bit Linux binary
make build-static
```

## Usage

### 1. Inspect Template Readiness
```bash
sudo ./bin/vm-template inspect
```

### 2. Generalize VM
```bash
# Dry run
sudo ./bin/vm-template prepare --dry-run

# Run sanitization and automatically power off
sudo ./bin/vm-template prepare --poweroff
```

### 3. Deploy to Remote VM
```bash
make deploy VM_HOST=192.168.1.89 VM_USER=vmware
```
