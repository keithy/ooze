package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		handleInit(args)
	case "send":
		handleSend(args)
	case "receive":
		handleReceive(args)
	case "list":
		handleList(args)
	case "switch":
		handleSwitch(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`ooze-nixos - NixOS root on ZFS mobility

Usage: ooze-nixos <command> [args...]

Commands:
  init<pool>               Initialize ZFS pool for NixOS root
  send <profile> <target>   Send NixOS config to target
  receive <pool>            Configure this host for NixOS root
  list                      List available NixOS configurations
  switch <profile>         Switch to profile (rebuild and switch)

Examples:
  ooze-nixos init tank
  ooze-nixos send default root@server2
  ooze-nixos receive tank
  ooze-nixos switch default
`)
}

func handleInit(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-nixos init <pool>")
		os.Exit(1)
	}
	pool := args[0]

	fmt.Printf("Initializing NixOS root on pool: %s\n", pool)

	execOrFatal("zfs", "create", "-o", "canmount=off", pool+"/os")
	execOrFatal("zfs", "create", "-o", "canmount=on", pool+"/os/nix")
	execOrFatal("zfs", "create", "-o", "canmount=on", pool+"/os/etc")
	execOrFatal("zfs", "create", "-o", "canmount=on", pool+"/os/var")
	execOrFatal("zfs", "create", "-o", "canmount=off", pool+"/safe")
	execOrFatal("zfs", "create", "-o", "canmount=on", "-o", "mountpoint=/home", pool+"/safe/user")
	execOrFatal("zfs", "create", "-o", "canmount=on", "-o", "mountpoint=/root", pool+"/safe/user/root")

	fmt.Println("NixOS ZFS layout created")
}

func handleSend(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-nixos send <profile> <target>")
		os.Exit(1)
	}
	profile := args[0]
	target := args[1]

	configPath := "/etc/nixos/configuration.nix"
	if profile != "default" {
		configPath = "/etc/nixos/profiles/" + profile + "/configuration.nix"
	}

	fmt.Printf("Sending NixOS config %s to %s\n", profile, target)

	tarCmd := exec.Command("tar", "-cf", "-", "-C", "/etc/nixos", ".")
	sshCmd := exec.Command("ssh", target, "tar", "-xf", "-", "-C", "/etc/nixos")

	tarCmd.Stdout, _ = sshCmd.StdinPipe()
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start ssh: %v\n", err)
		os.Exit(1)
	}
	if err := tarCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Tar failed: %v\n", err)
		os.Exit(1)
	}
	if err := sshCmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Transfer failed: %v\n", err)
		os.Exit(1)
	}

	_ = configPath // used for path resolution above
	fmt.Println("NixOS config sent")
}

func handleReceive(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-nixos receive <pool>")
		os.Exit(1)
	}
	pool := args[0]

	fmt.Printf("Receiving NixOS configuration into pool: %s\n", pool)

	execOrFatal("mkdir", "-p", "/mnt/etc/nixos")
	
	recvCmd := exec.Command("tar", "-xf", "-", "-C", "/mnt/etc/nixos")
	recvCmd.Stdin = os.Stdin
	recvCmd.Stdout = os.Stdout
	recvCmd.Stderr = os.Stderr

	if err := recvCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to receive config: %v\n", err)
		os.Exit(1)
	}

	sedPool := exec.Command("sed", "s/POOL/"+pool+"/g", "/mnt/etc/nixos/configuration.nix")
	sedPool.Stdout, _ = os.Create("/mnt/etc/nixos/configuration.nix")
	sedPool.Run()

	fmt.Println("NixOS config received. Run 'nixos-rebuild switch' to apply.")
}

func handleList(args []string) {
	profiles := []string{"default"}

	if _, err := os.Stat("/etc/nixos/profiles"); err == nil {
		entries, _ := os.ReadDir("/etc/nixos/profiles")
		for _, e := range entries {
			if e.IsDir() {
				profiles = append(profiles, e.Name())
			}
		}
	}

	fmt.Println("Available NixOS profiles:")
	for _, p := range profiles {
		fmt.Printf("  %s\n", p)
	}
}

func handleSwitch(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-nixos switch <profile>")
		os.Exit(1)
	}
	profile := args[0]

	if profile != "default" {
		configPath := "/etc/nixos/profiles/" + profile + "/configuration.nix"
		if _, err := os.Stat(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Profile not found: %s\n", profile)
			os.Exit(1)
		}
		execOrFatal("ln", "-sf", configPath, "/etc/nixos/configuration.nix")
	}

	cmd := exec.Command("nixos-rebuild", "switch")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Rebuild failed: %v\n", err)
		os.Exit(1)
	}
}

func execOrFatal(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		os.Exit(1)
	}
}

func replacePool(config string, pool string) string {
	return strings.ReplaceAll(config, "POOL", pool)
}
