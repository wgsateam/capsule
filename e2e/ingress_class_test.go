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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/clastix/capsule/api/v1alpha1"
)

var _ = Describe("when Tenant handles Ingress classes", func() {
	t := &v1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-class",
		},
		Spec: v1alpha1.TenantSpec{
			Owner: "ingress",
			StorageClasses: []string{},
			IngressClasses: []string{
				"nginx",
				"haproxy",
			},
			LimitRanges: []corev1.LimitRangeSpec{},
			NamespaceQuota: 3,
			NodeSelector:   map[string]string{},
			NetworkPolicies: []networkingv1.NetworkPolicySpec{},
			ResourceQuota: []corev1.ResourceQuotaSpec{},
		},
	}
	JustBeforeEach(func() {
		t.ResourceVersion = ""
		Expect(k8sClient.Create(context.TODO(), t)).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), t)).Should(Succeed())
	})
	It("should block non allowed Ingress class", func() {
		ns := NewNamespace("ingress-class-disallowed")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		By("non-specifying the class", func() {
			Eventually(func() (err error) {
				i := &v1beta12.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "denied-ingress",
					},
					Spec: v1beta12.IngressSpec{
						Backend: &v1beta12.IngressBackend{
							ServiceName: "foo",
							ServicePort: intstr.FromInt(8080),
						},
					},
				}
				_, err = cs.ExtensionsV1beta1().Ingresses(ns.GetName()).Create(context.TODO(), i, metav1.CreateOptions{})
				return
			}, 30*time.Second, time.Second).ShouldNot(Succeed())
		})
		By("specifying a forbidden class", func() {
			Eventually(func() (err error) {
				i := &v1beta12.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "denied-ingress",
					},
					Spec: v1beta12.IngressSpec{
						IngressClassName: pointer.StringPtr("the-worst-ingress-available"),
						Backend: &v1beta12.IngressBackend{
							ServiceName: "foo",
							ServicePort: intstr.FromInt(8080),
						},
					},
				}
				_, err = cs.ExtensionsV1beta1().Ingresses(ns.GetName()).Create(context.TODO(), i, metav1.CreateOptions{})
				return
			}, 30*time.Second, time.Second).ShouldNot(Succeed())
		})
	})
	It("should allow enabled Ingress class", func() {
		ns := NewNamespace("ingress-class-allowed")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		By("specifying using an available class", func() {
			for _, c := range t.Spec.IngressClasses {
				Eventually(func() (err error) {
					i := &v1beta12.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: c,
						},
						Spec: v1beta12.IngressSpec{
							IngressClassName: pointer.StringPtr(c),
							Backend: &v1beta12.IngressBackend{
								ServiceName: "foo",
								ServicePort: intstr.FromInt(8080),
							},
						},
					}
					_, err = cs.ExtensionsV1beta1().Ingresses(ns.GetName()).Create(context.TODO(), i, metav1.CreateOptions{})
					return
				}, 30*time.Second, time.Second).Should(Succeed())
			}
		})
	})
})
