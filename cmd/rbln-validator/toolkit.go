package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	toolkitReadyFile = "toolkit-ready"
	cdiRootPath      = "/var/run/cdi"
)

func newToolkitCommand(builder *configBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   "toolkit",
		Short: "Validate toolkit readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := builder.finalize()
			if err != nil {
				return err
			}
			return validateToolkit(cfg)
		},
	}
}

func validateToolkit(cfg *config) error {
	statusFile := filepath.Join(cfg.outputDir, toolkitReadyFile)
	if err := deleteStatusFile(statusFile); err != nil {
		return err
	}
	if err := ensureOutputDir(cfg.outputDir); err != nil {
		return err
	}

	checkReady := func() error {
		driverReadyPath := filepath.Join(cfg.outputDir, driverReadyFile)
		driverReadyInfo, err := os.Stat(driverReadyPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", driverReadyPath, err)
		}
		slog.Info(
			"toolkit validation: driver-ready mtime",
			"path",
			driverReadyPath,
			"mtime",
			driverReadyInfo.ModTime().Format(time.RFC3339),
		)

		entries, err := os.ReadDir(cdiRootPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", cdiRootPath, err)
		}
		slog.Info("toolkit validation: scanning CDI directory", "path", cdiRootPath, "entries", len(entries))
		var newestSpecPath string
		var newestSpecTime time.Time
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.Contains(strings.ToLower(name), "rbln") {
				path := filepath.Join(cdiRootPath, name)
				info, err := entry.Info()
				if err != nil {
					return fmt.Errorf("failed to stat %s: %w", path, err)
				}
				if info.Size() == 0 {
					slog.Info("toolkit validation: skipping empty CDI spec", "path", path)
					continue
				}
				if newestSpecPath == "" || info.ModTime().After(newestSpecTime) {
					newestSpecPath = path
					newestSpecTime = info.ModTime()
				}
			}
		}
		if newestSpecPath == "" {
			return fmt.Errorf("no rbln cdi spec found in %s", cdiRootPath)
		}
		slog.Info("toolkit validation: newest CDI spec", "path", newestSpecPath, "mtime", newestSpecTime.Format(time.RFC3339))
		if newestSpecTime.Before(driverReadyInfo.ModTime()) {
			slog.Warn(
				"toolkit validation: rbln cdi spec is stale; continuing",
				"specPath",
				newestSpecPath,
				"specMtime",
				newestSpecTime.Format(time.RFC3339),
				"driverReady",
				driverReadyPath,
				"driverReadyMtime",
				driverReadyInfo.ModTime().Format(time.RFC3339),
			)
			return nil
		}
		return nil
	}

	if cfg.withWait {
		for {
			if err := checkReady(); err != nil {
				slog.Info("toolkit is not ready", "err", err, "sleepSeconds", cfg.sleepIntervalSeconds)
				time.Sleep(time.Duration(cfg.sleepIntervalSeconds) * time.Second)
				continue
			}
			break
		}
	} else {
		if err := checkReady(); err != nil {
			slog.Error("toolkit is not ready", "err", err)
			return err
		}
	}

	return recreateStatusFile(cfg.outputDir, toolkitReadyFile)
}
