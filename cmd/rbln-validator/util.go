package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ensureOutputDir(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

func deleteStatusFile(statusFile string) error {
	if err := os.Remove(statusFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("remove status file %s: %w", statusFile, err)
		}
	}
	return nil
}

func recreateStatusFile(dir string, filename string) error {
	statusFile := filepath.Join(dir, filename)
	if err := os.Remove(statusFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove status file: %w", err)
	}
	// #nosec G304 -- status file path is controlled by operator config and output dir.
	file, err := os.Create(statusFile)
	if err != nil {
		return fmt.Errorf("create status file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close status file: %w", err)
	}
	return nil
}

func createStatusFileWithContent(statusFile string, content string) error {
	dir := filepath.Dir(statusFile)
	tmpFile, err := os.CreateTemp(dir, filepath.Base(statusFile)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary status file: %w", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temporary status file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary status file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if err := os.Rename(tmpFile.Name(), statusFile); err != nil {
		return fmt.Errorf("error moving temporary file to '%s': %w", statusFile, err)
	}
	return nil
}

func runCommand(name string, args []string, silent bool) error {
	cmd := exec.Command(name, args...)
	if !silent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func runCommandWithWait(name string, args []string, sleepSeconds int, silent bool) error {
	for {
		err := runCommand(name, args, silent)
		if err == nil {
			return nil
		}
		slog.Info("command failed, retrying", "command", name, "err", err)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}
