package sysprep

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultGenericNetplan(t *testing.T) {
	cfg := DefaultGenericNetplan()

	if cfg.Network.Version != 2 {
		t.Fatalf("expected network version 2, got %d", cfg.Network.Version)
	}

	eth, exists := cfg.Network.Ethernets["all-en"]
	if !exists {
		t.Fatalf("expected 'all-en' ethernet configuration to exist")
	}

	if !eth.DHCP4 {
		t.Errorf("expected dhcp4 to be true")
	}

	if eth.DHCPIdentifier != "mac" {
		t.Errorf("expected dhcp-identifier to be 'mac', got '%s'", eth.DHCPIdentifier)
	}

	if eth.Match["name"] != "en*" {
		t.Errorf("expected match name 'en*', got '%s'", eth.Match["name"])
	}
}

func TestWriteNetplanConfig(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "etc", "netplan", "00-installer-config.yaml")

	cfg := DefaultGenericNetplan()
	err := WriteNetplanConfig(targetFile, cfg)
	if err != nil {
		t.Fatalf("WriteNetplanConfig failed: %v", err)
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("target file was not created: %v", err)
	}

	// Verify Netplan strict permission requirements (0600)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file mode 0600, got %o", perm)
	}

	// Verify YAML content roundtrip
	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var parsed NetplanConfig
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("written file is not valid YAML: %v", err)
	}

	if parsed.Network.Version != 2 {
		t.Errorf("parsed version mismatch: %d", parsed.Network.Version)
	}
}
