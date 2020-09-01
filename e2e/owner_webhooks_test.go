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
	tnt := &v1alpha1.Tenant{
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
		tnt.ResourceVersion = ""
		Expect(k8sClient.Create(context.TODO(), tnt)).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), tnt)).Should(Succeed())
	})
	It("should disallow deletions", func() {
		By("blocking Capsule Limit ranges", func() {
			ns := NewNamespace("limit-range-disallow")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			lr := &corev1.LimitRange{}
			Eventually(func() error {
				n := fmt.Sprintf("capsule-%s-0", tnt.GetName())
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, lr)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

			cs := ownerClient(tnt)
			Expect(cs.CoreV1().LimitRanges(ns.GetName()).Delete(context.TODO(), lr.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
		})
		By("blocking Capsule Network Policy", func() {
			ns := NewNamespace("network-policy-disallow")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			np := &networkingv1.NetworkPolicy{}
			Eventually(func() error {
				n := fmt.Sprintf("capsule-%s-0", tnt.GetName())
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, np)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

			cs := ownerClient(tnt)
			Expect(cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
		})
		By("blocking blocking Capsule Resource Quota", func() {
			ns := NewNamespace("resource-quota-disallow")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			rq := &corev1.ResourceQuota{}
			Eventually(func() error {
				n := fmt.Sprintf("capsule-%s-0", tnt.GetName())
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns.GetName()}, rq)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

			cs := ownerClient(tnt)
			Expect(cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), rq.Name, metav1.DeleteOptions{})).ShouldNot(Succeed())
		})
	})
	It("should allow listing", func() {
		By("Limit Range resources", func() {
			ns := NewNamespace("limit-range-list")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			Eventually(func() (err error) {
				cs := ownerClient(tnt)
				_, err = cs.CoreV1().LimitRanges(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
				return
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
		By("Network Policy resources", func() {
			ns := NewNamespace("network-policy-list")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			Eventually(func() (err error) {
				cs := ownerClient(tnt)
				_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
				return
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
		By("Resource Quota resources", func() {
			ns := NewNamespace("resource-quota-list")
			NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
			NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

			Eventually(func() (err error) {
				cs := ownerClient(tnt)
				_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).List(context.TODO(), metav1.ListOptions{})
				return
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
	})
	It("should allow all actions to Tenant owner Network Policy resources", func() {
		ns := NewNamespace("network-policy-allow")
		NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
		NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)

		cs := ownerClient(tnt)
		np := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "custom-network-policy",
			},
			Spec: tnt.Spec.NetworkPolicies[0],
		}
		By("creating", func() {
			Eventually(func() (err error) {
				_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).Create(context.TODO(), np, metav1.CreateOptions{})
				return
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
		By("updating", func() {
			Eventually(func() (err error) {
				np.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{}
				_, err = cs.NetworkingV1().NetworkPolicies(ns.GetName()).Update(context.TODO(), np, metav1.UpdateOptions{})
				return
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
		By("deleting", func() {
			Eventually(func() (err error) {
				return cs.NetworkingV1().NetworkPolicies(ns.GetName()).Delete(context.TODO(), np.Name, metav1.DeleteOptions{})
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})
	})
})
