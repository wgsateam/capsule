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
	"context"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type RBACManager struct {
	ClientConfig *rest.Config
	CapsuleGroup string
	Log          logr.Logger
}

func (r RBACManager) SetupCapsuleRoles() error {
	// Create client as mgr.Client won't be availiable for us until we call mgr.Start
	k8sClient, err := client.New(r.ClientConfig, client.Options{})
	if err != nil {
		r.Log.Error(err, "Unable to create k8s client")
		return err
	}
	for roleName, role := range clusterRoles {
		r.Log.Info("setting up ClusterRoles", "ClusterRole", roleName)
		clusterRole := &rbacv1.ClusterRole{}
		if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: roleName}, clusterRole); err != nil {
			if errors.IsNotFound(err) {
				clusterRole.ObjectMeta = role.ObjectMeta
			}
		}
		_, err = controllerutil.CreateOrUpdate(context.TODO(), k8sClient, clusterRole, func() error {
			clusterRole.Rules = role.Rules
			return nil
		})
		if err != nil {
			return err
		}
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	provisionerClusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     r.CapsuleGroup,
	}
	r.Log.Info("setting up ClusterRoleBindings", "ClusterRoleBinding", provisionerRoleName)
	if err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: provisionerRoleName}, clusterRoleBinding); err != nil {
		if errors.IsNotFound(err) {
			if err = k8sClient.Create(context.TODO(), provisionerClusterRoleBinding); err != nil {
				return err
			}
		}
	} else {
		// RoleRef is immutable, so we need to delete and recreate ClusterRoleBinding if it changed
		if !equality.Semantic.DeepDerivative(provisionerClusterRoleBinding.RoleRef, clusterRoleBinding.RoleRef) {
			if err = k8sClient.Delete(context.TODO(), clusterRoleBinding); err != nil {
				return err
			}
		}
		_, err = controllerutil.CreateOrUpdate(context.TODO(), k8sClient, clusterRoleBinding, func() error {
			clusterRoleBinding.Subjects = provisionerClusterRoleBinding.Subjects
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
