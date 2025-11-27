package k8sutil

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

type RoleBuilder struct {
	*OwnableBuilder[rbacv1.Role, *rbacv1.Role]
}

func NewRoleBuilder(name, namespace string) *RoleBuilder {
	return &RoleBuilder{
		OwnableBuilder: NewOwnableBuilder[rbacv1.Role](name, namespace),
	}
}

func (b *RoleBuilder) WithRules(rules ...rbacv1.PolicyRule) *RoleBuilder {
	b.obj.Rules = append(b.obj.Rules, rules...)
	return b
}
