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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rebellionsaiv1alpha1 "github.com/rebellions-sw/rbln-npu-operator/api/v1alpha1"
	rblnv1beta1 "github.com/rebellions-sw/rbln-npu-operator/api/v1beta1"
)

var _ = Describe("RBLNDriver Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		rblndriver := &rebellionsaiv1alpha1.RBLNDriver{}
		clusterPolicyName := "cluster-policy"
		clusterPolicyKey := types.NamespacedName{
			Name: clusterPolicyName,
		}
		nodeName := "test-node"
		nodeKey := types.NamespacedName{
			Name: nodeName,
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind RBLNClusterPolicy")
			clusterPolicy := &rblnv1beta1.RBLNClusterPolicy{}
			err := k8sClient.Get(ctx, clusterPolicyKey, clusterPolicy)
			if err != nil && errors.IsNotFound(err) {
				resource := &rblnv1beta1.RBLNClusterPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: clusterPolicyName,
					},
					Spec: rblnv1beta1.RBLNClusterPolicySpec{
						Namespace:    "default",
						WorkloadType: "container",
						VFIOManager: rblnv1beta1.RBLNVFIOManagerSpec{
							Enabled: false,
						},
						SandboxDevicePlugin: rblnv1beta1.RBLNSandboxDevicePluginSpec{
							Enabled: false,
						},
						DevicePlugin: rblnv1beta1.RBLNDevicePluginSpec{
							Enabled: false,
						},
						MetricsExporter: rblnv1beta1.RBLNMetricsExporterSpec{
							Enabled: false,
						},
						NPUFeatureDiscovery: rblnv1beta1.RBLNNPUFeatureDiscoverySpec{
							Enabled: false,
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating a node that matches the driver selector")
			node := &corev1.Node{}
			err = k8sClient.Get(ctx, nodeKey, node)
			if err != nil && errors.IsNotFound(err) {
				resource := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName,
						Labels: map[string]string{
							"rebellions.ai/npu.present":                               "true",
							"feature.node.kubernetes.io/system-os_release.ID":         "ubuntu",
							"feature.node.kubernetes.io/system-os_release.VERSION_ID": "22.04",
							"feature.node.kubernetes.io/kernel-version.full":          "5.15.0-100-generic",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating the custom resource for the Kind RBLNDriver")
			err = k8sClient.Get(ctx, typeNamespacedName, rblndriver)
			if err != nil && errors.IsNotFound(err) {
				resource := &rebellionsaiv1alpha1.RBLNDriver{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: rebellionsaiv1alpha1.RBLNDriverSpec{
						Version: "3.0.0",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &rebellionsaiv1alpha1.RBLNDriver{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance RBLNDriver")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the specific resource instance RBLNClusterPolicy")
			policy := &rblnv1beta1.RBLNClusterPolicy{}
			err = k8sClient.Get(ctx, clusterPolicyKey, policy)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, policy)).To(Succeed())

			By("Cleanup the specific node instance")
			node := &corev1.Node{}
			err = k8sClient.Get(ctx, nodeKey, node)
			if err == nil {
				Expect(k8sClient.Delete(ctx, node)).To(Succeed())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RBLNDriverReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
