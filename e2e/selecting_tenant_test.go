//+build e2e

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

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/clastix/capsule/api/v1alpha1"
)

var _ = Describe("creating a Namespace with Tenant selector", func() {
	t1 := &v1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-one",
		},
		Spec: v1alpha1.TenantSpec{
			Owner:          "john",
			StorageClasses: []string{},
			IngressClasses: []string{},
			LimitRanges:    []corev1.LimitRangeSpec{},
			NamespaceQuota: 10,
			NodeSelector:   map[string]string{},
			ResourceQuota:  []corev1.ResourceQuotaSpec{},
		},
	}
	t2 := &v1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-two",
		},
		Spec: v1alpha1.TenantSpec{
			Owner:          "john",
			StorageClasses: []string{},
			IngressClasses: []string{},
			LimitRanges:    []corev1.LimitRangeSpec{},
			NamespaceQuota: 10,
			NodeSelector:   map[string]string{},
			ResourceQuota:  []corev1.ResourceQuotaSpec{},
		},
	}
	JustBeforeEach(func() {
		Expect(k8sClient.Create(context.TODO(), t1)).Should(Succeed())
		Expect(k8sClient.Create(context.TODO(), t2)).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), t1)).Should(Succeed())
		Expect(k8sClient.Delete(context.TODO(), t2)).Should(Succeed())
	})
	It("should be assigned to the selected Tenant", func() {
		ns := NewNamespace("tenant-2-ns")
		By("assigning to the Namespace the Capsule Tenant label", func() {
			l, err := v1alpha1.GetTypeLabel(&v1alpha1.Tenant{})
			Expect(err).ToNot(HaveOccurred())
			ns.Labels = map[string]string{
				l: t2.Name,
			}
		})
		NamespaceCreationShouldSucceed(ns, t2, defaultTimeoutInterval)
		NamespaceShouldBeManagedByTenant(ns, t2, defaultTimeoutInterval)
	})
})
