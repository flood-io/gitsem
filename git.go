package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runCommand(command ...string) error {
	cmd := exec.Command(command[0], command[1:]...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run command: %s\noutput:\n%s", err, output)
	}
	return nil
}

func isRepoClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	result := &bytes.Buffer{}
	cmd.Stdout = result
	if err := cmd.Run(); err != nil {
		return false, err
	}
	return result.String() == "", nil
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	result := &bytes.Buffer{}
	cmd.Stdout = result
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.String()), nil
}

func addFile(path string) error {
	return runCommand("git", "add", path)
}

func commit(message string) error {
	return runCommand("git", "commit", "-m", message)
}

func tag(version string) error {
	return runCommand("git", "tag", version)
}
