package sysprep

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dummy subdirectories and files
	subDir := filepath.Join(tmpDir, "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}

	dummyFile1 := filepath.Join(tmpDir, "file1.txt")
	dummyFile2 := filepath.Join(subDir, "file2.txt")

	_ = os.WriteFile(dummyFile1, []byte("test1"), 0644)
	_ = os.WriteFile(dummyFile2, []byte("test2"), 0644)

	cleanDir(tmpDir)

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read tmpDir: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected directory to be empty, found %d items", len(entries))
	}
}

func TestInspectionReport_Print(t *testing.T) {
	report := &InspectionReport{
		MachineIDLength: 0,
		SSHHostKeys:     nil,
		HasCloudInit:    true,
		HasOpenVMTools:  true,
		NetplanFiles:    []string{"/etc/netplan/00-installer.yaml"},
	}

	var buf bytes.Buffer
	report.Print(&buf)
	out := buf.String()

	if !bytes.Contains(buf.Bytes(), []byte("[✓] Machine ID: Truncated")) {
		t.Errorf("expected clean machine-id check in output: %s", out)
	}

	if !bytes.Contains(buf.Bytes(), []byte("[✓] SSH Host Keys: None found")) {
		t.Errorf("expected clean SSH host keys check in output: %s", out)
	}

	if !bytes.Contains(buf.Bytes(), []byte("[✓] VMware Tools: Installed")) {
		t.Errorf("expected VMware tools check in output: %s", out)
	}
}
