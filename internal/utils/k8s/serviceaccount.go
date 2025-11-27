package k8sutil

import (
	corev1 "k8s.io/api/core/v1"
)

type ServiceAccountBuilder struct {
	*OwnableBuilder[corev1.ServiceAccount, *corev1.ServiceAccount]
}

func NewServiceAccountBuilder(name, namespace string) *ServiceAccountBuilder {
	return &ServiceAccountBuilder{
		OwnableBuilder: NewOwnableBuilder[corev1.ServiceAccount](name, namespace),
	}
}
