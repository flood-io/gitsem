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

func runCommandStdout(command ...string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run command: %s\nstderr:\n%s\nstdout:\n%s", err, stderr, stdout)
	}
	return stdout.String(), nil
}

func runCommandTrimmed(command ...string) (string, error) {
	output, err := runCommandStdout(command...)

	return strings.TrimSpace(output), err
}

func isRepoClean() (bool, error) {
	result, err := runCommandTrimmed("git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return result == "", nil
}

func repoRoot() (string, error) {
	return runCommandTrimmed("git", "rev-parse", "--show-toplevel")
}

func sha() (string, error) {
	return runCommandTrimmed("git", "rev-parse", "HEAD")
}

func resetSHA(sha string) error {
	return runCommand("git", "reset", "--hard", sha)
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
