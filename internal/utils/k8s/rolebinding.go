package k8sutil

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

type RoleBindingBuilder struct {
	*OwnableBuilder[rbacv1.RoleBinding, *rbacv1.RoleBinding]
}

func NewRoleBindingBuilder(name, namespace string) *RoleBindingBuilder {
	return &RoleBindingBuilder{
		OwnableBuilder: NewOwnableBuilder[rbacv1.RoleBinding](name, namespace),
	}
}

func (b *RoleBindingBuilder) WithRoleRef(roleRef rbacv1.RoleRef) *RoleBindingBuilder {
	b.obj.RoleRef = roleRef
	return b
}

func (b *RoleBindingBuilder) WithSubjects(subjects ...rbacv1.Subject) *RoleBindingBuilder {
	b.obj.Subjects = append(b.obj.Subjects, subjects...)
	return b
}
