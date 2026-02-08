package controller

import (
	"context"
	"fmt"
	"strings"

	ocpconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/rebellions-sw/rbln-npu-operator/internal/consts"
)

type ClusterInfo struct {
	OpenshiftVersion string
	ContainerRuntime string
}

func NewClusterInfo(ctx context.Context, config *rest.Config) (*ClusterInfo, error) {
	ci := &ClusterInfo{}

	openshiftVersion, err := getOpenshiftVersion(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get openshift version: %w", err)
	}
	ci.OpenshiftVersion = openshiftVersion

	containerRuntime, err := getContainerRuntime(ctx, config, openshiftVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get container runtime: %w", err)
	}
	ci.ContainerRuntime = containerRuntime

	return ci, nil
}

func getOpenshiftVersion(ctx context.Context, config *rest.Config) (string, error) {
	client, err := ocpconfigv1.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("create openshift config client: %w", err)
	}

	v, err := client.ClusterVersions().Get(ctx, "version", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// not an OpenShift cluster
			return "", nil
		}
		return "", err
	}

	for _, condition := range v.Status.History {
		if condition.State != "Completed" {
			continue
		}

		ocpV := strings.Split(condition.Version, ".")
		if len(ocpV) > 1 {
			return ocpV[0] + "." + ocpV[1], nil
		}
		return ocpV[0], nil
	}

	return "", fmt.Errorf("failed to find Completed Openshift Cluster Version")
}

func getContainerRuntime(ctx context.Context, config *rest.Config, openshiftVersion string) (string, error) {
	if openshiftVersion != "" {
		return consts.CRIO, nil
	}

	k8sClient, err := corev1client.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("build k8s core v1 client: %w", err)
	}

	nodeSelector := labels.Set{consts.RBLNPresentLabelKey: "true"}.AsSelector().String()
	nodeList, err := k8sClient.Nodes().List(ctx, metav1.ListOptions{LabelSelector: nodeSelector})
	if err != nil {
		return "", fmt.Errorf("list nodes prior to checking container runtime: %w", err)
	}
	if len(nodeList.Items) == 0 {
		nodeList, err = k8sClient.Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", fmt.Errorf("list nodes for container runtime fallback: %w", err)
		}
	}

	var runtime string
	for _, node := range nodeList.Items {
		rt, err := getRuntimeString(node)
		if err != nil {
			continue
		}
		runtime = rt
		if runtime == consts.Containerd {
			break
		}
	}

	if runtime == "" {
		runtime = consts.Containerd
	}
	return runtime, nil
}

func getRuntimeString(node corev1.Node) (string, error) {
	runtimeVer := node.Status.NodeInfo.ContainerRuntimeVersion
	switch {
	case strings.HasPrefix(runtimeVer, "docker"):
		return consts.Docker, nil
	case strings.HasPrefix(runtimeVer, "containerd"):
		return consts.Containerd, nil
	case strings.HasPrefix(runtimeVer, "cri-o"):
		return consts.CRIO, nil
	default:
		return "", fmt.Errorf("runtime not recognized: %s", runtimeVer)
	}
}
