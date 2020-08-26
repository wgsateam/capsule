/*
Copyright 2020 Clastix Labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbac

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	provisionerRoleName = "capsule-namespace:provisioner"
	deleterRoleName     = "capsule-namespace:deleter"
)

var (
	clusterRoles = map[string]*rbacv1.ClusterRole{
		provisionerRoleName: {
			ObjectMeta: metav1.ObjectMeta{
				Name: provisionerRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces"},
					Verbs:     []string{"create"},
				},
			},
		},
		deleterRoleName: {
			ObjectMeta: metav1.ObjectMeta{
				Name: deleterRoleName,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces"},
					Verbs:     []string{"delete"},
				},
			},
		},
	}

	provisionerClusterRoleBinding = &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: provisionerRoleName,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     provisionerRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
)
