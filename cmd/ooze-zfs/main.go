package main

import (
	"encoding/json"
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
	case "send":
		handleSend(args)
	case "receive":
		handleReceive(args)
	case "list":
		handleList(args)
	case "snapshot":
		handleSnapshot(args)
	case "status":
		handleStatus(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`ooze-zfs - ZFS filesystem mobility

Usage: ooze-zfs <command> [args...]

Commands:
  send<dataset> <target> Send dataset to target server
 receive <dataset>          Receive dataset on this server
  list [dataset]             List datasets or snapshots
  snapshot <dataset>         Create migration snapshot
  status <dataset>           Show migration status

Examples:
  ooze-zfs send tank/data/db1 root@server2
  ooze-zfs receive tank/data/db1
  ooze-zfs snapshot tank/data/db1
  ooze-zfs list tank/data
`)
}

func handleSend(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-zfs send <dataset> <target>")
		os.Exit(1)
	}
	dataset := args[0]
	target := args[1]

	snapshot := dataset + "@ooze-migrate"
	
	execOrFatal("zfs", "snapshot", "-r", snapshot)

	sendCmd := exec.Command("zfs", "send", "-R", "-v", snapshot)
	recvCmd := exec.Command("ssh", target, "zfs", "receive", "-F", "-v", dataset)

	sendCmd.Stdout = os.Stdout
	sendCmd.Stderr = os.Stderr
	recvCmd.Stdin, _ = sendCmd.StdoutPipe()

	recvCmd.Stdout = os.Stdout
	recvCmd.Stderr = os.Stderr

	fmt.Printf("Sending %s to %s\n", snapshot, target)
	
	if err := recvCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start receive: %v\n", err)
		os.Exit(1)
	}
	if err := sendCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Send failed: %v\n", err)
		recvCmd.Process.Kill()
		os.Exit(1)
	}
	if err := recvCmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Receive failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration complete")
}

func handleReceive(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-zfs receive <dataset>")
		os.Exit(1)
	}
	dataset := args[0]

	recvCmd := exec.Command("zfs", "receive", "-F", "-v", dataset)
	recvCmd.Stdin = os.Stdin
	recvCmd.Stdout = os.Stdout
	recvCmd.Stderr = os.Stderr

	fmt.Printf("Receiving into %s\n", dataset)
	if err := recvCmd.Run(); err != nil {
		os.Exit(1)
	}
}

func handleList(args []string) {
	cmd := exec.Command("zfs", "list", "-t", "filesystem,snapshot", "-r", "-o", "name,used,refer,mountpoint")
	if len(args) > 0 {
		cmd = exec.Command("zfs", "list", "-t", "filesystem,snapshot", "-r", "-o", "name,used,refer,mountpoint", args[0])
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func handleSnapshot(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-zfs snapshot <dataset>")
		os.Exit(1)
	}
	dataset := args[0]
	snapshot := dataset + "@ooze-$(date +%Y%m%d-%H%M%S)"

	execOrFatal("zfs", "snapshot", "-r", snapshot)
	fmt.Printf("Created snapshot: %s\n", snapshot)
}

func handleStatus(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ooze-zfs status <dataset>")
		os.Exit(1)
	}
	dataset := args[0]

	cmd := exec.Command("zfs", "list", "-t", "snapshot", "-r", "-o", "name,creation,used", dataset)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
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

type MigrationStatus struct {
	Dataset   string `json:"dataset"`
	Snapshot  string `json:"snapshot"`
	SentSize  int64  `json:"sent_size"`
	Timestamp string `json:"timestamp"`
}

func readStatus(dataset string) (*MigrationStatus, error) {
	statusFile := "/var/lib/ooze/zfs/" + strings.ReplaceAll(dataset, "/", "_") + ".json"
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return nil, err
	}
	var status MigrationStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
