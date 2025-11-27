package k8sutil

import (
	corev1 "k8s.io/api/core/v1"
)

type ServiceBuilder struct {
	*OwnableBuilder[corev1.Service, *corev1.Service]
}

func NewServiceBuilder(name, namespace string) *ServiceBuilder {
	return &ServiceBuilder{
		OwnableBuilder: NewOwnableBuilder[corev1.Service](name, namespace),
	}
}

func (b *ServiceBuilder) WithAnnotations(annotations map[string]string) *ServiceBuilder {
	b.obj.Annotations = annotations
	return b
}

func (b *ServiceBuilder) WithLabels(labels map[string]string) *ServiceBuilder {
	b.obj.Labels = labels
	return b
}

func (b *ServiceBuilder) WithSelector(selector map[string]string) *ServiceBuilder {
	b.obj.Spec.Selector = selector
	return b
}

func (b *ServiceBuilder) WithPorts(ports []corev1.ServicePort) *ServiceBuilder {
	b.obj.Spec.Ports = ports
	return b
}
