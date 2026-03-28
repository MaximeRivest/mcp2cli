package expose

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Create ensures an exposed command exists in binDir and points to target.
func Create(binDir, name, target string) (string, error) {
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", fmt.Errorf("create expose bin directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		return createWindowsShim(binDir, name, target)
	}

	return createUnixShim(binDir, name, target)
}

// Remove deletes an exposed command shim if it exists.
func Remove(binDir, name string) error {
	if runtime.GOOS == "windows" {
		cmdPath := filepath.Join(binDir, name+".cmd")
		if err := removeIfExists(cmdPath); err != nil {
			return err
		}
		ps1Path := filepath.Join(binDir, name+".ps1")
		if err := removeIfExists(ps1Path); err != nil {
			return err
		}
		return nil
	}

	return removeIfExists(filepath.Join(binDir, name))
}

func createUnixShim(binDir, name, target string) (string, error) {
	path := filepath.Join(binDir, name)
	if err := removeIfExists(path); err != nil {
		return "", err
	}

	if err := os.Symlink(target, path); err == nil {
		return path, nil
	}

	wrapper := fmt.Sprintf("#!/usr/bin/env sh\nexec %q \"$@\"\n", target)
	if err := os.WriteFile(path, []byte(wrapper), 0o755); err != nil {
		return "", fmt.Errorf("write wrapper %s: %w", path, err)
	}
	return path, nil
}

func createWindowsShim(binDir, name, target string) (string, error) {
	cmdPath := filepath.Join(binDir, name+".cmd")
	cmdContents := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", target)
	if err := os.WriteFile(cmdPath, []byte(cmdContents), 0o644); err != nil {
		return "", fmt.Errorf("write wrapper %s: %w", cmdPath, err)
	}

	ps1Path := filepath.Join(binDir, name+".ps1")
	ps1Contents := fmt.Sprintf("& %q @args\r\n", target)
	if err := os.WriteFile(ps1Path, []byte(ps1Contents), 0o644); err != nil {
		return "", fmt.Errorf("write wrapper %s: %w", ps1Path, err)
	}

	return cmdPath, nil
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}
