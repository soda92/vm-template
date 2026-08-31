package sysprep

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// InspectionReport summarizes template readiness.
type InspectionReport struct {
	MachineIDLength int64
	SSHHostKeys     []string
	HasCloudInit    bool
	HasOpenVMTools  bool
	NetplanFiles    []string
}

// Inspect collects status information from the host.
func Inspect() (*InspectionReport, error) {
	report := &InspectionReport{}

	// Machine ID
	if info, err := os.Stat("/etc/machine-id"); err == nil {
		report.MachineIDLength = info.Size()
	} else {
		report.MachineIDLength = -1
	}

	// SSH Host keys
	keys, _ := filepath.Glob("/etc/ssh/ssh_host_*")
	report.SSHHostKeys = keys

	// Cloud-init
	if _, err := exec.LookPath("cloud-init"); err == nil {
		report.HasCloudInit = true
	}

	// open-vm-tools
	if _, err := exec.LookPath("vmtoolsd"); err == nil {
		report.HasOpenVMTools = true
	}

	// Netplan configs
	configs, _ := filepath.Glob("/etc/netplan/*")
	report.NetplanFiles = configs

	return report, nil
}

// Print formats and displays the inspection report.
func (r *InspectionReport) Print(w io.Writer) {
	fmt.Fprintf(w, "=== VM Template State Inspection ===\n")
	if r.MachineIDLength == 0 {
		fmt.Fprintf(w, "  [✓] Machine ID: Truncated (Ready for regeneration)\n")
	} else if r.MachineIDLength > 0 {
		fmt.Fprintf(w, "  [!] Machine ID: Populated (%d bytes - NEEDS CLEANUP)\n", r.MachineIDLength)
	} else {
		fmt.Fprintf(w, "  [!] Machine ID: Missing (Should be truncated to 0 bytes)\n")
	}

	if len(r.SSHHostKeys) == 0 {
		fmt.Fprintf(w, "  [✓] SSH Host Keys: None found (Ready for regeneration)\n")
	} else {
		fmt.Fprintf(w, "  [!] SSH Host Keys: %d keys present (NEEDS CLEANUP)\n", len(r.SSHHostKeys))
		for _, k := range r.SSHHostKeys {
			fmt.Fprintf(w, "      - %s\n", k)
		}
	}

	if r.HasOpenVMTools {
		fmt.Fprintf(w, "  [✓] VMware Tools: Installed (vmtoolsd detected)\n")
	} else {
		fmt.Fprintf(w, "  [!] VMware Tools: Not detected (open-vm-tools recommended)\n")
	}

	if r.HasCloudInit {
		fmt.Fprintf(w, "  [*] Cloud-Init: Installed\n")
	} else {
		fmt.Fprintf(w, "  [*] Cloud-Init: Not installed (using standard guest customization)\n")
	}

	fmt.Fprintf(w, "  [*] Netplan Configs: %d file(s)\n", len(r.NetplanFiles))
	for _, n := range r.NetplanFiles {
		fmt.Fprintf(w, "      - %s\n", n)
	}
	fmt.Fprintf(w, "====================================\n")
}
