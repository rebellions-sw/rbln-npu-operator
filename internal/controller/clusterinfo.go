package controller

import (
	"context"
	"fmt"
	"strings"

	ocpconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type ClusterInfo struct {
	OpenshiftVersion string
}

func NewClusterInfo(ctx context.Context, config *rest.Config) (*ClusterInfo, error) {
	ci := &ClusterInfo{}

	openshiftVersion, err := getOpenshiftVersion(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get openshift version: %w", err)
	}
	ci.OpenshiftVersion = openshiftVersion

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
