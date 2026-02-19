package k8sutil

import (
	corev1 "k8s.io/api/core/v1"
)

type PodSpecBuilder struct {
	*Builder[corev1.PodSpec]
}

func NewPodSpecBuilder(v ...*corev1.PodSpec) *PodSpecBuilder {
	podSpec := &corev1.PodSpec{}
	if len(v) > 0 {
		podSpec = v[0]
	}
	return &PodSpecBuilder{Builder: NewBuilder(podSpec)}
}

func (b *PodSpecBuilder) WithAffinity(affinity *corev1.Affinity) *PodSpecBuilder {
	b.obj.Affinity = affinity
	return b
}

func (b *PodSpecBuilder) WithTolerations(tolerations []corev1.Toleration) *PodSpecBuilder {
	b.obj.Tolerations = tolerations
	return b
}

func (b *PodSpecBuilder) WithImagePullSecrets(secrets []string) *PodSpecBuilder {
	imgPullSecrets := []corev1.LocalObjectReference{}
	for _, secret := range secrets {
		imgPullSecrets = append(imgPullSecrets, corev1.LocalObjectReference{Name: secret})
	}
	b.obj.ImagePullSecrets = imgPullSecrets
	return b
}

func (b *PodSpecBuilder) WithPriorityClassName(priorityClass string) *PodSpecBuilder {
	b.obj.PriorityClassName = priorityClass
	return b
}

func (b *PodSpecBuilder) WithVolumes(volumes []corev1.Volume) *PodSpecBuilder {
	b.obj.Volumes = volumes
	return b
}

func (b *PodSpecBuilder) WithNodeSelector(selector map[string]string) *PodSpecBuilder {
	b.obj.NodeSelector = selector
	return b
}

func (b *PodSpecBuilder) WithContainers(containers []*corev1.Container) *PodSpecBuilder {
	b.obj.Containers = make([]corev1.Container, len(containers))
	for idx, cont := range containers {
		b.obj.Containers[idx] = *cont
	}
	return b
}

func (b *PodSpecBuilder) WithInitContainers(containers []*corev1.Container) *PodSpecBuilder {
	b.obj.InitContainers = make([]corev1.Container, len(containers))
	for idx, cont := range containers {
		b.obj.InitContainers[idx] = *cont
	}
	return b
}

func (b *PodSpecBuilder) WithHostNetwork(enabled bool) *PodSpecBuilder {
	b.obj.HostNetwork = enabled
	return b
}

func (b *PodSpecBuilder) WithHostPID(enabled bool) *PodSpecBuilder {
	b.obj.HostPID = enabled
	return b
}

func (b *PodSpecBuilder) WithTerminationGracePeriodSeconds(seconds int64) *PodSpecBuilder {
	b.obj.TerminationGracePeriodSeconds = &[]int64{seconds}[0]
	return b
}

func (b *PodSpecBuilder) WithServiceAccountName(serviceAccountName string) *PodSpecBuilder {
	b.obj.ServiceAccountName = serviceAccountName
	return b
}

// MergeAffinity combines default and user-provided Affinity
func MergeAffinity(defaultAffinity, userAffinity *corev1.Affinity) *corev1.Affinity {
	if userAffinity == nil {
		return defaultAffinity
	}
	if defaultAffinity == nil {
		return userAffinity
	}

	merged := &corev1.Affinity{}

	// Merge NodeAffinity
	if userAffinity.NodeAffinity != nil {
		merged.NodeAffinity = userAffinity.NodeAffinity
	} else {
		merged.NodeAffinity = defaultAffinity.NodeAffinity
	}
	if merged.NodeAffinity != nil && defaultAffinity.NodeAffinity != nil && userAffinity.NodeAffinity != nil {
		// Merge RequiredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
				merged.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = defaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(
					merged.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
					defaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms...,
				)
			}
		}
		// Merge PreferredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
				merged.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = defaultAffinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
					merged.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					defaultAffinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution...,
				)
			}
		}
	}

	// Merge PodAffinity
	if userAffinity.PodAffinity != nil {
		merged.PodAffinity = userAffinity.PodAffinity
	} else {
		merged.PodAffinity = defaultAffinity.PodAffinity
	}
	if merged.PodAffinity != nil && defaultAffinity.PodAffinity != nil && userAffinity.PodAffinity != nil {
		// Merge RequiredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
				merged.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = defaultAffinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
					merged.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
					defaultAffinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution...,
				)
			}
		}
		// Merge PreferredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
				merged.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = defaultAffinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
					merged.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					defaultAffinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution...,
				)
			}
		}
	}

	// Merge PodAntiAffinity
	if userAffinity.PodAntiAffinity != nil {
		merged.PodAntiAffinity = userAffinity.PodAntiAffinity
	} else {
		merged.PodAntiAffinity = defaultAffinity.PodAntiAffinity
	}
	if merged.PodAntiAffinity != nil && defaultAffinity.PodAntiAffinity != nil && userAffinity.PodAntiAffinity != nil {
		// Merge RequiredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
				merged.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = defaultAffinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
					merged.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
					defaultAffinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...,
				)
			}
		}
		// Merge PreferredDuringSchedulingIgnoredDuringExecution
		if defaultAffinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
			if merged.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
				merged.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = defaultAffinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			} else {
				merged.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
					merged.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					defaultAffinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution...,
				)
			}
		}
	}

	return merged
}
