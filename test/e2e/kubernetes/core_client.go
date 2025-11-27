package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type CoreClient struct {
	k8sClient corev1client.CoreV1Interface
}

func NewClient(k8sClient corev1client.CoreV1Interface) *CoreClient {
	return &CoreClient{
		k8sClient: k8sClient,
	}
}

func (cc *CoreClient) CreateNamespace(
	ctx context.Context,
	namespaceName string,
	labels map[string]string,
) (*corev1.Namespace, error) {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: labels,
		},
		Status: corev1.NamespaceStatus{},
	}

	return cc.k8sClient.Namespaces().Create(ctx, namespaceObj, metav1.CreateOptions{})
}

func (cc *CoreClient) GetPodsByLabel(
	ctx context.Context,
	namespace string,
	labelMap map[string]string,
) ([]corev1.Pod, error) {
	podList, err := cc.k8sClient.Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (cc *CoreClient) IsPodReady(ctx context.Context, podName, namespace string) (bool, error) {
	pod, err := cc.k8sClient.Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("unexpected error getting pod  %s: %w", podName, err)
	}

	for _, c := range pod.Status.Conditions {
		if c.Type != corev1.PodReady {
			continue
		}
		if c.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}

func (cc *CoreClient) ListNodes(ctx context.Context, selector map[string]string) ([]corev1.Node, error) {
	opts := metav1.ListOptions{}
	if len(selector) > 0 {
		opts.LabelSelector = labels.SelectorFromSet(selector).String()
	}
	nodes, err := cc.k8sClient.Nodes().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

func (cc *CoreClient) IsNodeReady(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (cc *CoreClient) DeleteNamespace(ctx context.Context, namespaceName string) error {
	return cc.k8sClient.Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
}
