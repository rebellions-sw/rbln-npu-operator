package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultOutputDir            = "/run/rbln/validations"
	defaultOperatorNamespace    = "rbln-system"
	defaultSleepIntervalSeconds = 5

	envOutputDir            = "OUTPUT_DIR"
	envNamespace            = "OPERATOR_NAMESPACE"
	envWithWait             = "WITH_WAIT"
	envSleepIntervalSeconds = "SLEEP_INTERVAL_SECONDS"
)

type config struct {
	namespace            string
	outputDir            string
	withWait             bool
	sleepIntervalSeconds int
}

type configBuilder struct {
	namespace            string
	outputDir            string
	withWait             bool
	sleepIntervalSeconds int
}

func newConfigBuilder() *configBuilder {
	return &configBuilder{
		namespace:            envString(envNamespace, defaultOperatorNamespace),
		outputDir:            envString(envOutputDir, defaultOutputDir),
		withWait:             envBool(envWithWait, false),
		sleepIntervalSeconds: envInt(envSleepIntervalSeconds, defaultSleepIntervalSeconds),
	}
}

func (b *configBuilder) bindFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVar(&b.namespace, "namespace", b.namespace, "namespace where operator resources run")
	flags.StringVar(&b.outputDir, "output-dir", b.outputDir, "output directory for validation status files")
	flags.BoolVar(&b.withWait, "with-wait", b.withWait, "wait for validation to complete successfully")
	flags.IntVar(
		&b.sleepIntervalSeconds,
		"sleep-interval-seconds",
		b.sleepIntervalSeconds,
		"sleep interval in seconds between retries",
	)
}

func (b *configBuilder) finalize() (*config, error) {
	outputDir := strings.TrimSpace(b.outputDir)
	if outputDir == "" {
		return nil, fmt.Errorf("output-dir must not be empty")
	}
	namespace := strings.TrimSpace(b.namespace)
	if namespace == "" {
		return nil, fmt.Errorf("namespace must not be empty")
	}
	if b.sleepIntervalSeconds <= 0 {
		return nil, fmt.Errorf("sleep-interval-seconds must be greater than 0")
	}

	return &config{
		namespace:            namespace,
		outputDir:            outputDir,
		withWait:             b.withWait,
		sleepIntervalSeconds: b.sleepIntervalSeconds,
	}, nil
}

func envString(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func envBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func envInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
