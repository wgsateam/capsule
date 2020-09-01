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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/clastix/capsule/api/v1alpha1"
)

var _ = Describe("changing Tenant managed Kubernetes resources", func() {
	tnt := &v1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-resources-changes",
		},
		Spec: v1alpha1.TenantSpec{
			Owner:          "laura",
			StorageClasses: []string{},
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
						{
							Type: corev1.LimitTypeContainer,
							Default: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
							DefaultRequest: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("10Mi"),
							},
							Min: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							},
							Max: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
						{
							Type: corev1.LimitTypePersistentVolumeClaim,
							Min: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
							Max: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
			NetworkPolicies: []networkingv1.NetworkPolicySpec{
				{
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							From: []networkingv1.NetworkPolicyPeer{
								{
									NamespaceSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"capsule.clastix.io/tenant": "tenant-resources",
										},
									},
								},
								{
									PodSelector: &metav1.LabelSelector{},
								},
								{
									IPBlock: &networkingv1.IPBlock{
										CIDR: "192.168.0.0/12",
									},
								},
							},
						},
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{
						{
							To: []networkingv1.NetworkPolicyPeer{
								{
									IPBlock: &networkingv1.IPBlock{
										CIDR: "0.0.0.0/0",
										Except: []string{
											"192.168.0.0/12",
										},
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
			NamespaceQuota: 4,
			NodeSelector: map[string]string{
				"kubernetes.io/os": "linux",
			},
			ResourceQuota: []corev1.ResourceQuotaSpec{
				{
					Hard: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceLimitsCPU:      resource.MustParse("8"),
						corev1.ResourceLimitsMemory:   resource.MustParse("16Gi"),
						corev1.ResourceRequestsCPU:    resource.MustParse("8"),
						corev1.ResourceRequestsMemory: resource.MustParse("16Gi"),
					},
					Scopes: []corev1.ResourceQuotaScope{
						corev1.ResourceQuotaScopeNotTerminating,
					},
				},
				{
					Hard: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourcePods: resource.MustParse("10"),
					},
				},
				{
					Hard: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceRequestsStorage: resource.MustParse("100Gi"),
					},
				},
			},
		},
	}
	nsl := []string{"fire", "walk", "with", "me"}
	JustBeforeEach(func() {
		tnt.ResourceVersion = ""
		Expect(k8sClient.Create(context.TODO(), tnt)).Should(Succeed())
		By("creating the Namespaces", func() {
			for _, i := range nsl {
				ns := NewNamespace(i)
				NamespaceCreationShouldSucceed(ns, tnt, defaultTimeoutInterval)
				NamespaceShouldBeManagedByTenant(ns, tnt, defaultTimeoutInterval)
			}
		})
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), tnt)).Should(Succeed())
	})
	It("should reapply the original resources upon third party change", func() {
		for _, ns := range nsl {
			By("changing Limit Range resources", func() {
				for i, s := range tnt.Spec.LimitRanges {
					n := fmt.Sprintf("capsule-%s-%d", tnt.GetName(), i)
					lr := &corev1.LimitRange{}
					Eventually(func() error {
						return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, lr)
					}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

					c := lr.DeepCopy()
					c.Spec.Limits = []corev1.LimitRangeItem{}
					Expect(k8sClient.Update(context.TODO(), c, &client.UpdateOptions{})).Should(Succeed())

					Eventually(func() corev1.LimitRangeSpec {
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, lr)).Should(Succeed())
						return lr.Spec
					}, defaultTimeoutInterval, defaultPollInterval).Should(Equal(s))
				}
			})
			By("changing Network Policy resources", func() {
				for i, s := range tnt.Spec.NetworkPolicies {
					n := fmt.Sprintf("capsule-%s-%d", tnt.GetName(), i)
					np := &networkingv1.NetworkPolicy{}
					Eventually(func() error {
						return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, np)
					}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
					Expect(np.Spec).Should(Equal(s))

					c := np.DeepCopy()
					c.Spec.Egress = []networkingv1.NetworkPolicyEgressRule{}
					c.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{}
					Expect(k8sClient.Update(context.TODO(), c, &client.UpdateOptions{})).Should(Succeed())

					Eventually(func() networkingv1.NetworkPolicySpec {
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, np)).Should(Succeed())
						return np.Spec
					}, defaultTimeoutInterval, defaultPollInterval).Should(Equal(s))
				}
			})
			By("changing Resource Quota resources", func() {
				for i, s := range tnt.Spec.ResourceQuota {
					n := fmt.Sprintf("capsule-%s-%d", tnt.GetName(), i)
					rq := &corev1.ResourceQuota{}
					Eventually(func() error {
						return k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, rq)
					}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

					c := rq.DeepCopy()
					c.Spec.Hard = map[corev1.ResourceName]resource.Quantity{}
					Expect(k8sClient.Update(context.TODO(), c, &client.UpdateOptions{})).Should(Succeed())

					Eventually(func() corev1.ResourceQuotaSpec {
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: ns}, rq)).Should(Succeed())
						return rq.Spec
					}, defaultTimeoutInterval, defaultPollInterval).Should(Equal(s))
				}
			})
		}
	})
})
