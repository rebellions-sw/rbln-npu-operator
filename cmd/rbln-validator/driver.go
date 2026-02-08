package main

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	driverContainerReadyFile   = ".driver-ctr-ready"
	driverReadyFile            = "driver-ready"
	driverContainerLibraryPath = "/usr/local/lib/rbln"
)

func newDriverCommand(builder *configBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   "driver",
		Short: "Validate driver readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := builder.finalize()
			if err != nil {
				return err
			}
			return validateDriver(cmd.Context(), cfg)
		},
	}
}

func validateDriver(ctx context.Context, cfg *config) error {
	if err := deleteStatusFile(filepath.Join(cfg.outputDir, driverReadyFile)); err != nil {
		return err
	}

	if err := ensureOutputDir(cfg.outputDir); err != nil {
		return err
	}

	driver := &Driver{
		outputDir:            cfg.outputDir,
		namespace:            cfg.namespace,
		withWait:             cfg.withWait,
		sleepIntervalSeconds: cfg.sleepIntervalSeconds,
		ctx:                  ctx,
	}
	driverInfo, err := driver.runValidation(false)
	if err != nil {
		slog.Error("driver is not ready", "err", err)
		return err
	}
	slog.Info("driver validation completed", "hostDriver", driverInfo.isHostDriver)

	return driver.createStatusFile(driverInfo)
}
