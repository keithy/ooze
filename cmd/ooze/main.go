package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		return
	}

	subcommand := os.Args[1]

	if subcommand == "--commands" {
		listCommands()
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}
	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	prefix := exeName
	if strings.HasPrefix(prefix, "ooze-") {
		prefix = "ooze"
	}
	prefix = strings.TrimSuffix(prefix, "-wrapper")

	subPath := filepath.Join(exeDir, subcommand)
	if _, err := os.Stat(subPath); err == nil {
		run(subPath, os.Args[2:])
		return
	}

	buildDir := filepath.Join(exeDir, "build")
	prefixedName := prefix + "-" + subcommand
	subPath = filepath.Join(buildDir, prefixedName)
	if _, err := os.Stat(subPath); err == nil {
		run(subPath, os.Args[2:])
		return
	}

	prefixedPath := filepath.Join(exeDir, prefixedName)
	if _, err := os.Stat(prefixedPath); err == nil {
		run(prefixedPath, os.Args[2:])
		return
	}

	siblingLib := filepath.Join(exeDir, "..", "lib", "ooze")
	prefixedPath = filepath.Join(siblingLib, prefixedName)
	if _, err := os.Stat(prefixedPath); err == nil {
		run(prefixedPath, os.Args[2:])
		return
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
	printUsage()
}

func run(path string, args []string) {
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	exePath, _ := os.Executable()
	exeName := filepath.Base(exePath)

	prefix := exeName
	if strings.HasPrefix(prefix, "ooze-") {
		prefix = "ooze"
	}
	prefix = strings.TrimSuffix(prefix, "-wrapper")
	prefixLen := len(prefix) + 1

	fmt.Printf("Usage: %s <command> [args...]\n", exeName)
	fmt.Println("")
	fmt.Println("Available commands:")

	exeDir := filepath.Dir(exePath)
	entries, err := os.ReadDir(exeDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == exeName {
				continue
			}
			if len(name) > prefixLen && name[:prefixLen] == prefix+"-" {
				fmt.Printf("  %s\n", name[prefixLen:])
			}
		}
	}

	fmt.Println("")
	fmt.Printf("Use '%s <command> --help' for command-specific help\n", exeName)
}

func listCommands() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[]")
		return
	}
	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	prefix := exeName
	if strings.HasPrefix(prefix, "ooze-") {
		prefix = "ooze"
	}
	prefix = strings.TrimSuffix(prefix, "-wrapper")
	prefixLen := len(prefix) + 1

	commands := []string{}
	entries, err := os.ReadDir(exeDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == exeName {
				continue
			}
			if len(name) > prefixLen && name[:prefixLen] == prefix+"-" {
				commands = append(commands, name[prefixLen:])
			}
		}
	}

	fmt.Print("[")
	for i, cmd := range commands {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("\"%s\"", cmd)
	}
	fmt.Println("]")
}
