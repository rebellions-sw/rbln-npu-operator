package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	driverManagerAppLabelKey = "app.kubernetes.io/component"
	driverManagerName        = "rbln-driver"
	rblnAPIGroupPrefix       = "rebellions.ai/"
	rblnClusterPolicyKind    = "RBLNClusterPolicy"
	rblnDriverKind           = "RBLNDriver"
	shell                    = "sh"
	hostRootDefault          = "/"
	hostRootMountPath        = "/host"
	hostUsrBinMountPath      = "/host-usr-bin"
	hostRblnSMIPath          = hostUsrBinMountPath + "/rbln-smi"
	hostDriverModuleName     = "rebellions"
	driverInstallDirDefault  = "/run/rbln/driver"
)

type Driver struct {
	outputDir            string
	namespace            string
	withWait             bool
	sleepIntervalSeconds int
	ctx                  context.Context
}

type driverInfo struct {
	isHostDriver         bool
	hostRoot             string
	driverRoot           string
	driverRootCtrPath    string
	containerLibraryPath string
}

func (d *Driver) runValidation(silent bool) (driverInfo, error) {
	if err := validateHostDriver(silent); err == nil {
		slog.Info("Detected a pre-installed driver on the host")
		return getDriverInfo(true, hostRootDefault, hostRootDefault, driverContainerLibraryPath), nil
	}

	if err := validateDriverContainer(
		d.ctx,
		d.namespace,
		d.outputDir,
		d.withWait,
		d.sleepIntervalSeconds,
		silent,
	); err != nil {
		return driverInfo{}, err
	}

	return getDriverInfo(false, hostRootDefault, driverInstallDirDefault, driverContainerLibraryPath), nil
}

func validateDriverContainer(
	ctx context.Context,
	namespace string,
	outputDir string,
	withWait bool,
	sleepIntervalSeconds int,
	silent bool,
) error {
	driverManagedByOperator, err := isDriverManagedByOperator(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error checking if driver is managed by operator: %w", err)
	}
	if driverManagedByOperator {
		slog.Info("Driver is not pre-installed on the host and is managed by operator. Checking driver container status.")
		if err := assertDriverContainerReady(outputDir, withWait, sleepIntervalSeconds, silent); err != nil {
			return fmt.Errorf("error checking driver container status: %w", err)
		}
	}

	driverRoot := driverRoot(driverInstallDirDefault)
	validateDriver := func(silent bool) error {
		driverLibraryPath, err := driverRoot.getDriverLibraryPath()
		if err != nil {
			return fmt.Errorf("failed to locate driver libraries: %w", err)
		}

		rblnSMIPath, err := driverRoot.getRblnSMIPath()
		if err != nil {
			return fmt.Errorf("failed to locate rbln-smi: %w", err)
		}
		slog.Info("using rbln-smi", "path", rblnSMIPath)

		cmd := exec.Command(rblnSMIPath)
		cmd.Env = setEnvVar(os.Environ(), "LD_PRELOAD", prependPathListEnvvar("LD_PRELOAD", driverLibraryPath))
		if !silent {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		return cmd.Run()
	}

	for {
		slog.Info("Attempting to validate a driver container installation")
		if err := validateDriver(silent); err != nil {
			if !withWait {
				return fmt.Errorf("error validating driver: %w", err)
			}
			slog.Info(
				"failed to validate the driver, retrying",
				"sleepSeconds",
				sleepIntervalSeconds,
				"err",
				err,
			)
			time.Sleep(time.Duration(sleepIntervalSeconds) * time.Second)
			continue
		}
		return nil
	}
}

func isDriverManagedByOperator(ctx context.Context, namespace string) (bool, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return false, fmt.Errorf("error getting cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return false, fmt.Errorf("error getting k8s client: %w", err)
	}

	selector := labels.Set{driverManagerAppLabelKey: driverManagerName}.AsSelector().String()
	opts := metav1.ListOptions{LabelSelector: selector}
	dsList, err := kubeClient.AppsV1().DaemonSets(namespace).List(ctx, opts)
	if err != nil {
		return false, fmt.Errorf("error listing daemonsets: %w", err)
	}
	slog.Info("driver manager discovery", "namespace", namespace, "daemonsets", len(dsList.Items))

	for i := range dsList.Items {
		ds := dsList.Items[i]
		owner := metav1.GetControllerOf(&ds)
		if owner == nil {
			continue
		}
		if strings.HasPrefix(owner.APIVersion, rblnAPIGroupPrefix) &&
			(owner.Kind == rblnClusterPolicyKind || owner.Kind == rblnDriverKind) {
			slog.Info("driver manager discovery: managed by operator", "daemonset", ds.Name)
			return true, nil
		}
	}

	slog.Info("driver manager discovery: not managed by operator", "namespace", namespace)
	return false, nil
}

func assertDriverContainerReady(outputDir string, withWait bool, sleepIntervalSeconds int, silent bool) error {
	readyPath := filepath.Join(outputDir, driverContainerReadyFile)
	args := []string{"-c", fmt.Sprintf("stat %s", readyPath)}
	if withWait {
		return runCommandWithWait(shell, args, sleepIntervalSeconds, silent)
	}
	return runCommand(shell, args, silent)
}

func validateHostDriver(silent bool) error {
	slog.Info("Attempting to validate a pre-installed driver on the host")
	installed, version, err := detectInstalledHostDriver()
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("host driver module %q not installed", hostDriverModuleName)
	}
	loaded, err := detectLoadedHostDriver()
	if err != nil {
		return err
	}
	if !loaded {
		return fmt.Errorf("host driver module %q not loaded", hostDriverModuleName)
	}
	if version != "" {
		slog.Info("Detected host driver module version", "module", hostDriverModuleName, "version", version)
	}
	fileInfo, err := os.Lstat(hostRblnSMIPath)
	if err != nil {
		return fmt.Errorf("no 'rbln-smi' file present on the host: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("empty 'rbln-smi' file found on the host")
	}
	return runCommand(hostRblnSMIPath, []string{}, silent)
}

func detectInstalledHostDriver() (bool, string, error) {
	versionCmd := exec.Command(
		"chroot",
		hostRootMountPath,
		"modinfo",
		"-F",
		"version",
		hostDriverModuleName,
	)
	versionOut, err := versionCmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, "", nil
		}
		slog.Debug("host driver version check failed", "err", err)
		return false, "", err
	}
	version := strings.TrimSpace(string(versionOut))
	if version != "" {
		slog.Info("Detected host driver module version", "module", hostDriverModuleName, "version", version)
	}
	return version != "", version, nil
}

func detectLoadedHostDriver() (bool, error) {
	loadedCmd := exec.Command(
		"chroot",
		hostRootMountPath,
		"test",
		"-d",
		"/sys/module/"+hostDriverModuleName,
	)
	if err := loadedCmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		slog.Debug("host driver load check failed", "err", err)
		return false, err
	}
	slog.Info("Detected host driver module loaded", "module", hostDriverModuleName)
	return true, nil
}

func getDriverInfo(
	isHostDriver bool,
	hostRoot string,
	driverInstallDir string,
	containerLibraryPath string,
) driverInfo {
	if isHostDriver {
		return driverInfo{
			isHostDriver:         true,
			hostRoot:             hostRoot,
			driverRoot:           hostRoot,
			driverRootCtrPath:    "/host",
			containerLibraryPath: containerLibraryPath,
		}
	}

	return driverInfo{
		isHostDriver:         false,
		hostRoot:             hostRoot,
		driverRoot:           driverInstallDir,
		driverRootCtrPath:    "/host",
		containerLibraryPath: containerLibraryPath,
	}
}

func (d *Driver) createStatusFile(info driverInfo) error {
	statusFileContent := strings.Join([]string{
		fmt.Sprintf("IS_HOST_DRIVER=%t", info.isHostDriver),
		fmt.Sprintf("RBLN_CTK_DAEMON_HOST_ROOT=%s", info.driverRootCtrPath),
		fmt.Sprintf("RBLN_CTK_DAEMON_DRIVER_ROOT=%s", info.driverRoot),
		fmt.Sprintf("RBLN_CTK_DAEMON_CONTAINER_LIBRARY_PATH=%s", info.containerLibraryPath),
	}, "\n") + "\n"

	return createStatusFileWithContent(filepath.Join(d.outputDir, driverReadyFile), statusFileContent)
}

func setEnvVar(envvars []string, key, value string) []string {
	updated := make([]string, 0, len(envvars)+1)
	for _, envvar := range envvars {
		pair := strings.SplitN(envvar, "=", 2)
		if pair[0] == key {
			continue
		}
		updated = append(updated, envvar)
	}
	return append(updated, fmt.Sprintf("%s=%s", key, value))
}

func prependPathListEnvvar(envvar string, prepend ...string) string {
	if len(prepend) == 0 {
		return os.Getenv(envvar)
	}
	current := filepath.SplitList(os.Getenv(envvar))
	return strings.Join(append(prepend, current...), string(filepath.ListSeparator))
}
