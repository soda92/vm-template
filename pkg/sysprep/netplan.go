package sysprep

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NetplanEthernet represents network interface ethernet config in Netplan.
type NetplanEthernet struct {
	DHCP4          bool              `yaml:"dhcp4,omitempty"`
	DHCP6          bool              `yaml:"dhcp6,omitempty"`
	DHCPIdentifier string            `yaml:"dhcp-identifier,omitempty"`
	Optional       bool              `yaml:"optional,omitempty"`
	Match          map[string]string `yaml:"match,omitempty"`
	SetName        string            `yaml:"set-name,omitempty"`
}

// NetplanConfig represents a Netplan version 2 YAML structure.
type NetplanConfig struct {
	Network struct {
		Version   int                        `yaml:"version"`
		Renderer  string                     `yaml:"renderer,omitempty"`
		Ethernets map[string]NetplanEthernet `yaml:"ethernets,omitempty"`
	} `yaml:"network"`
}

// DefaultGenericNetplan generates a universal Netplan config matching any ethernet interface.
func DefaultGenericNetplan() *NetplanConfig {
	cfg := &NetplanConfig{}
	cfg.Network.Version = 2
	cfg.Network.Ethernets = map[string]NetplanEthernet{
		"all-en": {
			Match: map[string]string{
				"name": "en*",
			},
			DHCP4:          true,
			DHCPIdentifier: "mac",
			Optional:       true,
		},
		"all-eth": {
			Match: map[string]string{
				"name": "eth*",
			},
			DHCP4:          true,
			DHCPIdentifier: "mac",
			Optional:       true,
		},
	}
	return cfg
}

// WriteNetplanConfig safely writes a netplan configuration file.
func WriteNetplanConfig(destPath string, cfg *NetplanConfig) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal netplan config: %w", err)
	}

	// Netplan configs must have restricted permissions (0600)
	if err := os.WriteFile(destPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write netplan config to %s: %w", destPath, err)
	}

	return nil
}
