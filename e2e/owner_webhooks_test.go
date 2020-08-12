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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/clastix/capsule/api/v1alpha1"
)

var _ = Describe("when Tenant owner interacts with the webhooks", func() {
	newNamespace := func(name string) (ns *corev1.Namespace) {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		return
	}

	t := &v1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-owner",
		},
		Spec: v1alpha1.TenantSpec{
			Owner: "ruby",
			StorageClasses: []string{
				"cephfs",
				"glusterfs",
			},
			IngressClasses: []string{},
			LimitRanges: []corev1.LimitRangeSpec{
				{
					Limits: []corev1.LimitRangeItem{
						{
							Type: corev1.LimitTypePod,
							Min: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							},
							Max: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			NamespaceQuota: 3,
			NodeSelector:   map[string]string{},
			NetworkPolicies: []networkingv1.NetworkPolicySpec{
				{
					Egress: []networkingv1.NetworkPolicyEgressRule{
						{
							To: []networkingv1.NetworkPolicyPeer{
								{
									IPBlock: &networkingv1.IPBlock{
										CIDR: "0.0.0.0/0",
									},
								},
							},
						},
					},
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
				},
			},
			ResourceQuota: []corev1.ResourceQuotaSpec{
				{
					Hard: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourcePods: resource.MustParse("10"),
					},
				},
			},
		},
	}
	JustBeforeEach(func() {
		t.ResourceVersion = ""
		Expect(k8sClient.Create(context.TODO(), t)).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), t)).Should(Succeed())
	})
	It("should disallow deletion of Capsule Limit Range", func() {
		ns := newNamespace("limit-range-disallow")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		lr := &corev1.LimitRange{}
		Eventually(func() error {
			n := fmt.Sprintf("capsule-%s-0", t.GetName())
			return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, lr)
		}, 30*time.Second, time.Second).Should(Succeed())

		Expect(cs.CoreV1().LimitRanges(ns.GetName()).Delete(context.TODO(), lr.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
	})
	It("should allow listing of Limit Range resources", func() {
		ns := newNamespace("limit-range-list")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		Eventually(func() (err error) {
			_, err = cs.CoreV1().LimitRanges(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
			return
		}, 30*time.Second, time.Second).Should(Succeed())
	})
	It("should disallow deletion of Capsule Network Policy", func() {
		ns := newNamespace("network-policy-disallow")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		np := &networkingv1.NetworkPolicy{}
		Eventually(func() error {
			n := fmt.Sprintf("capsule-%s-0", t.GetName())
			return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, np)
		}, 20*time.Second, time.Second).Should(Succeed())

		Expect(cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
	})
	It("should allow listing of Network Policy resources", func() {
		ns := newNamespace("network-policy-list")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		Eventually(func() (err error) {
			_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
			return
		}, 30*time.Second, time.Second).Should(Succeed())
	})
	It("should allow all actions to Tenant owner Network Policy resources", func() {
		ns := newNamespace("network-policy-allow")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		np := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "custom-network-policy",
			},
			Spec: t.Spec.NetworkPolicies[0],
		}
		Eventually(func() (err error) {
			_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).Create(context.TODO(), np, metav1.CreateOptions{})
			return
		}, 30*time.Second, time.Second).Should(Succeed())
		Eventually(func() (err error) {
			np.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{}
			_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).Update(context.TODO(), np, metav1.UpdateOptions{})
			return
		}, 30*time.Second, time.Second).Should(Succeed())
		Eventually(func() (err error) {
			return cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})
		}, 30*time.Second, time.Second).Should(Succeed())
	})

	It("should disallow deletion of Capsule Resource Quota", func() {
		ns := newNamespace("resource-quota-disallow")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		rq := &corev1.ResourceQuota{}
		Eventually(func() error {
			n := fmt.Sprintf("capsule-%s-0", t.GetName())
			return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, rq)
		}, 30*time.Second, time.Second).Should(Succeed())

		Expect(cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), rq.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
	})
	It("should allow listing of Resource Quota resources", func() {
		ns := newNamespace("resource-quota-list")
		cs := ownerClient(t)

		NamespaceCreationShouldSucceed(ns, t)
		NamespaceShouldBeManagedByTenant(ns, t)

		Eventually(func() (err error) {
			_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
			return
		}, 30*time.Second, time.Second).Should(Succeed())
	})
})
