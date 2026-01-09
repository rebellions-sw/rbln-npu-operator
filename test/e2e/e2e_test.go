/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	e2ek8s "github.com/rebellions-sw/rbln-npu-operator/test/e2e/kubernetes"
	e2elog "github.com/rebellions-sw/rbln-npu-operator/test/e2e/logs"
	"github.com/rebellions-sw/rbln-npu-operator/test/e2e/testenv"
)

const (
	defaultOperandPollInterval = 5 * time.Second
	defaultOperandWaitTimeout  = 15 * time.Minute
	NPUResourceName            = corev1.ResourceName("rebellions.ai/ATOM")
	devicePluginNodeLabelKey   = "rebellions.ai/npu.deploy.device-plugin"
	devicePluginNodeLabelValue = "true"
	rblnClusterPolicyCRDName   = "rblnclusterpolicies.rebellions.ai"
)

var _ = Describe("e2e-npu-operator-scenario-test", Ordered, func() {
	te := testenv.NewTestEnv("rbln-npu-operator")

	Describe("NPU Operator RBLNClusterPolicy", func() {
		Context("Container-type NPU Operator deployment", Ordered, func() {
			/*
			   Scenario:
			   - Deploy NPU Operator CRD with container-type configuration.
			   - Verify that all operator-managed components (device-plugin, feature-discovery, metrics-exporter.)
			     are successfully created, running, and reporting healthy states.
			   - Validate CR status conditions, DaemonSet/Pod readiness, and functional behavior.
			*/

			var (
				helmClient          *HelmClient
				helmReleaseName     string
				k8sCoreClient       *e2ek8s.CoreClient
				k8sExtensionsClient *e2ek8s.ExtensionClient
				testNamespace       *corev1.Namespace
			)

			BeforeAll(func(ctx context.Context) {
				var err error
				k8sCoreClient = e2ek8s.NewClient(te.ClientSet.CoreV1())
				k8sExtensionsClient = e2ek8s.NewExtensionClient(te.ExtClientSet)
				nsLabels := map[string]string{
					"e2e-run": string(testenv.RunID),
				}

				testNamespace, err = k8sCoreClient.CreateNamespace(ctx, e2eCfg.namespace, nsLabels)
				if err != nil {
					Fail(fmt.Sprintf("failed to create gpu operator namespace %s: %v", e2eCfg.namespace, err))
				}

				helmClient, err = NewHelmClient(
					testNamespace.Name,
					testenv.TestCtx.KubeConfig,
					e2eCfg.helmChart,
				)
				if err != nil {
					Fail(fmt.Sprintf("failed to instantiate gpu operator client: %v", err))
				}

				helmReleaseName, err = helmClient.Install(ctx, ChartOptions{
					CleanupOnFail: true,
					ReleaseName:   "rbln-npu-operator",
					Timeout:       5 * time.Minute,
					Wait:          true,
					Values: map[string]interface{}{
						"operator": map[string]interface{}{
							"image": map[string]interface{}{
								"registry":   e2eCfg.operatorRegistry,
								"repository": e2eCfg.operatorRepository,
								"tag":        e2eCfg.operatorVersion,
							},
						},
					},
				})

				e2elog.Infof("HelmReleaseName: %s", helmReleaseName)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterAll(func(ctx context.Context) {
				err := helmClient.Uninstall(ctx, helmReleaseName)
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}

				err = k8sExtensionsClient.DeleteCRD(ctx, rblnClusterPolicyCRDName)
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}

				err = k8sCoreClient.DeleteNamespace(ctx, testNamespace.Name)
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should bring up the all of the operand pods successfully", func(ctx context.Context) {
				operands := []string{
					"rbln-device-plugin",
					"rbln-metrics-exporter",
					"rbln-npu-feature-discovery",
				}
				e2elog.Infof("Ensure that the npu operator operands come up")
				for _, operand := range operands {
					By(fmt.Sprintf("waiting for %s pods to become ready", operand))
					Eventually(func() bool {
						labelMap := map[string]string{
							"app": operand,
						}
						pods, err := k8sCoreClient.GetPodsByLabel(ctx, testNamespace.Name, labelMap)
						if err != nil {
							e2elog.Infof("WARN: error retrieving pods of operand %s: %v", operand, err)
							return false
						}

						var readyCount int
						for _, pod := range pods {
							e2elog.Infof("Checking status of pod %s", pod.Name)
							isReady, err := k8sCoreClient.IsPodReady(ctx, pod.Name, pod.Namespace)
							if err != nil {
								e2elog.Infof("WARN: error when retrieving pod status of %s/%s: %v", testNamespace.Name, operand, err)
								return false
							}
							if isReady {
								readyCount++
							}
						}
						return len(pods) > 0 && readyCount == len(pods)
					}).WithPolling(defaultOperandPollInterval).Within(defaultOperandWaitTimeout).WithContext(ctx).Should(BeTrue())
				}
			})
			It("should advertise rebellions.ai/ATOM on ready nodes", func(ctx context.Context) {
				Eventually(func(g Gomega) bool {
					nodes, err := k8sCoreClient.ListNodes(ctx, map[string]string{
						devicePluginNodeLabelKey: devicePluginNodeLabelValue,
					})
					g.Expect(err).NotTo(HaveOccurred())

					found := false
					for i := range nodes {
						node := &nodes[i]
						if !k8sCoreClient.IsNodeReady(node) {
							continue
						}

						allocQty, allocOK := node.Status.Allocatable[NPUResourceName]
						if !allocOK || allocQty.Value() == 0 {
							e2elog.Infof("node %s has no allocatable %s", node.Name, NPUResourceName)
							return false
						}

						capQty, capOK := node.Status.Capacity[NPUResourceName]
						if !capOK || capQty.Value() == 0 {
							e2elog.Infof("node %s has no capacity %s", node.Name, NPUResourceName)
							return false
						}

						found = true
					}
					return found
				}).WithContext(ctx).
					WithPolling(defaultOperandPollInterval).
					Within(defaultOperandWaitTimeout).
					Should(BeTrue(), "no ready labeled node exposed rebellions.ai/ATOM")
			})
		})
	})
	print(te)
})
