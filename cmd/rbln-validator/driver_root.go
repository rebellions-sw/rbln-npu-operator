package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type driverRoot string

func (r driverRoot) getDriverLibraryPath() (string, error) {
	librarySearchPaths := []string{
		"/usr/lib",
		"/usr/lib64",
		"/usr/lib/x86_64-linux-gnu",
		"/usr/lib/aarch64-linux-gnu",
		"/lib64",
		"/lib/x86_64-linux-gnu",
		"/lib/aarch64-linux-gnu",
	}

	libraryPath, err := r.findFile("librbln-ml.so", librarySearchPaths...)
	if err != nil {
		return "", err
	}

	return libraryPath, nil
}

func (r driverRoot) getRblnSMIPath() (string, error) {
	binarySearchPaths := []string{
		"/usr/bin",
		"/usr/sbin",
		"/bin",
		"/sbin",
	}

	binaryPath, err := r.findFile("rbln-smi", binarySearchPaths...)
	if err != nil {
		return "", err
	}

	return binaryPath, nil
}

func (r driverRoot) findFile(name string, searchIn ...string) (string, error) {
	for _, d := range append([]string{""}, searchIn...) {
		relative := strings.TrimPrefix(d, "/")
		l := filepath.Join(string(r), relative, name)
		candidate, err := resolveLink(l)
		if err != nil {
			continue
		}
		return candidate, nil
	}

	return "", fmt.Errorf("error locating %q", name)
}

func resolveLink(l string) (string, error) {
	resolved, err := filepath.EvalSymlinks(l)
	if err != nil {
		return "", fmt.Errorf("error resolving link '%s': %w", l, err)
	}
	return resolved, nil
}
