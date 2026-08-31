package sysprep

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckRootStep verifies the process is executing with effective UID 0.
type CheckRootStep struct{}

func (s *CheckRootStep) Name() string { return "Verifying root privileges" }
func (s *CheckRootStep) Run(ctx *Context) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this utility must be run as root (UID 0)")
	}
	return nil
}

// StopServicesStep stops logging services so file handles on logs can be cleanly released.
type StopServicesStep struct{}

func (s *StopServicesStep) Name() string { return "Stopping journald and rsyslog services" }
func (s *StopServicesStep) Run(ctx *Context) error {
	_ = exec.Command("systemctl", "stop", "systemd-journald.service", "rsyslog.service").Run()
	return nil
}

// CloudInitCleanStep cleans cloud-init metadata and artifacts if installed.
type CloudInitCleanStep struct{}

func (s *CloudInitCleanStep) Name() string { return "Resetting cloud-init state" }
func (s *CloudInitCleanStep) Run(ctx *Context) error {
	if _, err := exec.LookPath("cloud-init"); err == nil {
		cmd := exec.Command("cloud-init", "clean", "--logs", "--seed")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cloud-init clean error: %s (%w)", string(out), err)
		}
	}
	return nil
}

// ResetMachineIDStep truncates /etc/machine-id to 0 bytes and fixes dbus symlink.
type ResetMachineIDStep struct{}

func (s *ResetMachineIDStep) Name() string { return "Truncating /etc/machine-id" }
func (s *ResetMachineIDStep) Run(ctx *Context) error {
	machineIDPath := "/etc/machine-id"
	// Truncate to 0 bytes so systemd regenerates it on next boot
	if err := os.Truncate(machineIDPath, 0); err != nil {
		if os.IsNotExist(err) {
			if f, err := os.Create(machineIDPath); err == nil {
				_ = f.Close()
			}
		} else {
			return fmt.Errorf("failed to truncate %s: %w", machineIDPath, err)
		}
	}

	dbusPath := "/var/lib/dbus/machine-id"
	_ = os.Remove(dbusPath)
	_ = os.MkdirAll(filepath.Dir(dbusPath), 0755)
	_ = os.Symlink(machineIDPath, dbusPath)
	return nil
}

// ResetSSHHostKeysStep removes all generated SSH host keys.
type ResetSSHHostKeysStep struct{}

func (s *ResetSSHHostKeysStep) Name() string { return "Removing existing SSH host keys" }
func (s *ResetSSHHostKeysStep) Run(ctx *Context) error {
	files, err := filepath.Glob("/etc/ssh/ssh_host_*")
	if err != nil {
		return fmt.Errorf("failed to find ssh host keys: %w", err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", f, err)
		}
	}
	return nil
}

// EnableSSHKeyRegenStep creates a systemd oneshot service to regenerate SSH host keys on first boot.
type EnableSSHKeyRegenStep struct{}

func (s *EnableSSHKeyRegenStep) Name() string { return "Ensuring SSH host key regeneration on first boot" }
func (s *EnableSSHKeyRegenStep) Run(ctx *Context) error {
	serviceContent := `[Unit]
Description=Regenerate SSH Host Keys
ConditionPathEmpty=/etc/ssh/ssh_host_rsa_key
Before=ssh.service

[Service]
Type=oneshot
ExecStart=/usr/bin/ssh-keygen -A
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`
	servicePath := "/etc/systemd/system/regenerate-ssh-host-keys.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", servicePath, err)
	}

	_ = exec.Command("systemctl", "daemon-reload").Run()
	_ = exec.Command("systemctl", "enable", "regenerate-ssh-host-keys.service").Run()
	return nil
}

// ConfigureNetplanStep writes a clean, MAC-independent Netplan configuration.
type ConfigureNetplanStep struct{}

func (s *ConfigureNetplanStep) Name() string { return "Configuring generic Netplan settings" }
func (s *ConfigureNetplanStep) Run(ctx *Context) error {
	if ctx.SkipNetplan {
		return nil
	}

	// Remove old Netplan YAML configs in /etc/netplan/
	existing, _ := filepath.Glob("/etc/netplan/*.yaml")
	for _, f := range existing {
		_ = os.Remove(f)
	}
	existingYml, _ := filepath.Glob("/etc/netplan/*.yml")
	for _, f := range existingYml {
		_ = os.Remove(f)
	}

	cfg := DefaultGenericNetplan()
	return WriteNetplanConfig("/etc/netplan/00-installer-config.yaml", cfg)
}

// CleanLogsAndCachesStep removes package caches, temporary files, and truncates log files.
type CleanLogsAndCachesStep struct{}

func (s *CleanLogsAndCachesStep) Name() string { return "Cleaning logs, package caches, and temporary files" }
func (s *CleanLogsAndCachesStep) Run(ctx *Context) error {
	// APT cache
	if _, err := exec.LookPath("apt-get"); err == nil {
		_ = exec.Command("apt-get", "autoremove", "-y", "--purge").Run()
		_ = exec.Command("apt-get", "clean").Run()
		_ = os.RemoveAll("/var/lib/apt/lists")
		_ = os.MkdirAll("/var/lib/apt/lists/partial", 0755)
	}

	// Truncate /var/log files
	_ = filepath.Walk("/var/log", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.HasSuffix(path, ".gz") {
			_ = os.Truncate(path, 0)
		}
		if err == nil && strings.HasSuffix(path, ".gz") {
			_ = os.Remove(path)
		}
		return nil
	})

	// Temporary dirs
	cleanDir("/tmp")
	cleanDir("/var/tmp")

	return nil
}

func cleanDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		_ = os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

// CleanShellHistoryStep removes bash/zsh history for all users and root.
type CleanShellHistoryStep struct{}

func (s *CleanShellHistoryStep) Name() string { return "Cleaning user shell history files" }
func (s *CleanShellHistoryStep) Run(ctx *Context) error {
	historyFiles := []string{
		"/root/.bash_history",
		"/root/.zsh_history",
	}

	homeEntries, _ := os.ReadDir("/home")
	for _, entry := range homeEntries {
		if entry.IsDir() {
			historyFiles = append(historyFiles,
				filepath.Join("/home", entry.Name(), ".bash_history"),
				filepath.Join("/home", entry.Name(), ".zsh_history"),
				filepath.Join("/home", entry.Name(), ".local/share/fish/fish_history"),
			)
		}
	}

	for _, f := range historyFiles {
		_ = os.Remove(f)
	}
	return nil
}

// PoweroffStep shuts down the system.
type PoweroffStep struct{}

func (s *PoweroffStep) Name() string { return "Shutting down system" }
func (s *PoweroffStep) Run(ctx *Context) error {
	if !ctx.PoweroffOnFinish {
		return nil
	}
	return exec.Command("poweroff").Run()
}
