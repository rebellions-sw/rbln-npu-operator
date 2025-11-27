package testenv

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var RunID = uuid.NewUUID()

type TestEnv struct {
	BaseName     string
	clientConfig *rest.Config
	ClientSet    clientset.Interface
}

func NewTestEnv(baseName string) *TestEnv {
	te := &TestEnv{
		BaseName: baseName,
	}

	ginkgo.BeforeEach(te.BeforeEach)

	return te
}

func (te *TestEnv) BeforeEach(ctx context.Context) {
	ginkgo.DeferCleanup(te.AfterEach)

	ginkgo.By("Creating a kubernetes client")
	cfg, err := LoadRESTClientConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	te.clientConfig = rest.CopyConfig(cfg)
	te.ClientSet, err = clientset.NewForConfig(cfg)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func (te *TestEnv) AfterEach(ctx context.Context) {
	defer func() {
		te.clientConfig = nil
		te.ClientSet = nil
	}()
}
