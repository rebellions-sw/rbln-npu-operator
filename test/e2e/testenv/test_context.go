package testenv

import (
	"flag"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

type TestContext struct {
	KubeConfig  string
	KubeContext string
}

var TestCtx = TestContext{}

func RegisterContextFlags(flags *flag.FlagSet) {
	flag.StringVar(
		&TestCtx.KubeConfig,
		clientcmd.RecommendedConfigPathFlag,
		os.Getenv(clientcmd.RecommendedConfigPathEnvVar),
		"Path to the kubeconfig file (defaults to $KUBECONFIG)",
	)
	flag.StringVar(
		&TestCtx.KubeContext,
		clientcmd.FlagContext,
		"",
		"kubecontext to use (defaults to current-context)",
	)
}
