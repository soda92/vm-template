package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/soda/vm-template/pkg/sysprep"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "inspect":
		runInspect(os.Args[2:])
	case "prepare", "sanitize":
		runPrepare(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`vm-template: Linux VM Template Preparation & Sanitization CLI for VMware vSphere

Usage:
  vm-template <command> [flags]

Commands:
  inspect    Check current VM template readiness (machine-id, keys, tools)
  prepare    Run full sanitization pipeline to prepare VM for template conversion
  help       Show help documentation

Flags for 'prepare':
  --dry-run         Simulate pipeline without modifying system files
  --poweroff        Automatically shutdown VM once preparation is complete
  --skip-netplan    Do not modify Netplan configuration
  --verbose         Enable verbose logging output`)
}

func runInspect(_ []string) {
	report, err := sysprep.Inspect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting VM: %v\n", err)
		os.Exit(1)
	}
	report.Print(os.Stdout)
}

func runPrepare(args []string) {
	fs := flag.NewFlagSet("prepare", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Simulate pipeline without modifying system files")
	poweroff := fs.Bool("poweroff", false, "Automatically shutdown VM once preparation is complete")
	skipNetplan := fs.Bool("skip-netplan", false, "Do not modify Netplan configuration")
	verbose := fs.Bool("verbose", false, "Enable verbose logging output")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	ctx := &sysprep.Context{
		DryRun:           *dryRun,
		PoweroffOnFinish: *poweroff,
		SkipNetplan:      *skipNetplan,
		Verbose:          *verbose,
		Out:              os.Stdout,
	}

	pipeline := sysprep.NewPipeline(
		&sysprep.CheckRootStep{},
		&sysprep.StopServicesStep{},
		&sysprep.CloudInitCleanStep{},
		&sysprep.ResetMachineIDStep{},
		&sysprep.ResetSSHHostKeysStep{},
		&sysprep.EnableSSHKeyRegenStep{},
		&sysprep.ConfigureNetplanStep{},
		&sysprep.CleanLogsAndCachesStep{},
		&sysprep.CleanShellHistoryStep{},
		&sysprep.PoweroffStep{},
	)

	fmt.Println("Starting VM Template Sanitization Pipeline...")
	if err := pipeline.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "\n[ERROR] Pipeline failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n[SUCCESS] VM is generalized and ready to be converted into a vSphere Template.")
	if !*poweroff {
		fmt.Println("Tip: Run 'sudo poweroff' before converting to template in vSphere.")
	}
}
