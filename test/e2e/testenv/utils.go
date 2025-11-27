package testenv

import (
	"fmt"
	"strings"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/onsi/ginkgo/v2"

	e2elog "github.com/rebellions-sw/rbln-npu-operator/test/e2e/logs"
)

func LoadRESTClientConfig() (config *restclient.Config, err error) {
	kubeCfg, err := loadKubeRESTConfig(TestCtx.KubeContext)
	if err != nil {
		if TestCtx.KubeConfig == "" {
			return restclient.InClusterConfig()
		}
		return nil, err
	}

	cfg, err := clientcmd.NewDefaultClientConfig(*kubeCfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}

	setSpecUserAgent(cfg)
	return cfg, nil
}

func setSpecUserAgent(cfg *restclient.Config) {
	spec := ginkgo.CurrentSpecReport()
	if len(spec.ContainerHierarchyTexts) == 0 && spec.LeafNodeText == "" {
		return
	}

	parts := append([]string{}, spec.ContainerHierarchyTexts...)
	if spec.LeafNodeText != "" {
		parts = append(parts, spec.LeafNodeText)
	}

	cfg.UserAgent = fmt.Sprintf("%s -- %s", restclient.DefaultKubernetesUserAgent(), strings.Join(parts, " "))
}

func loadKubeRESTConfig(kubeContext string) (*clientcmdapi.Config, error) {
	e2elog.Infof("kubeConfig path: %s", TestCtx.KubeConfig)
	if TestCtx.KubeConfig == "" {
		return nil, fmt.Errorf("KubeConfig path must be specified")
	}
	cfg, err := clientcmd.LoadFromFile(TestCtx.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("loading KubeConfig: %w", err)
	}
	if kubeContext != "" {
		e2elog.Infof("overriding kubeContext: %s", kubeContext)
		cfg.CurrentContext = kubeContext
	}
	return cfg, nil
}
