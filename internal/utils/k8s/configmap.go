package k8sutil

import (
	corev1 "k8s.io/api/core/v1"
)

type ConfigMapBuilder struct {
	*OwnableBuilder[corev1.ConfigMap, *corev1.ConfigMap]
}

func NewConfigMapBuilder(name, namespace string) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		OwnableBuilder: NewOwnableBuilder[corev1.ConfigMap](name, namespace),
	}
}

func (b *ConfigMapBuilder) WithData(data map[string]string) *ConfigMapBuilder {
	b.obj.Data = data
	return b
}
