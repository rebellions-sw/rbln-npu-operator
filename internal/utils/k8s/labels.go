package k8sutil

import (
	"strings"

	"github.com/rebellions-sw/rbln-npu-operator/internal/consts"
)

// IsControlPlane return true if node is a control plane node
func IsControlPlane(labels map[string]string) bool {
	_, isCP := labels["node-role.kubernetes.io/control-plane"]
	return isCP
}

// IsNfdProvisioned return true if node labels contain NFD labels
func IsNfdProvisioned(labels map[string]string) bool {
	for key := range labels {
		if strings.HasPrefix(key, consts.NFDLabelPrefix) {
			return true
		}
	}
	return false
}

// IsRblnNode return true if node has Rebellions device
func IsRblnNode(labels map[string]string) bool {
	if _, ok := labels[consts.RBLNPresentLabelKey]; ok {
		return labels[consts.RBLNPresentLabelKey] == "true"
	}
	return false
}

// GetWorkloadType returns the workload type from labels or the default workload if not found
func GetWorkloadType(labels map[string]string, defaultWorkload string) string {
	if workloadType, ok := labels[consts.RBLNWorkloadConfigLabelKey]; ok {
		return workloadType
	}
	return defaultWorkload
}
