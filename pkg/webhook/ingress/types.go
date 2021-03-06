// Copyright 2020-2021 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package ingress

import (
	"sort"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

const (
	annotationName = "kubernetes.io/ingress.class"
)

type Ingress interface {
	IngressClass() *string
	Namespace() string
	Name() string
	Hostnames() []string
}

type NetworkingV1 struct {
	*networkingv1.Ingress
}

func (n NetworkingV1) Name() string {
	return n.GetName()
}

func (n NetworkingV1) IngressClass() (res *string) {
	res = n.Spec.IngressClassName
	if res == nil {
		if a := n.GetAnnotations(); a != nil {
			if v, ok := a[annotationName]; ok {
				res = &v
			}
		}
	}
	return
}

func (n NetworkingV1) Namespace() string {
	return n.GetNamespace()
}

func (n NetworkingV1) Hostnames() []string {
	rules := n.Spec.Rules
	var hostnames []string
	for _, el := range rules {
		hostnames = append(hostnames, el.Host)
	}
	return hostnames
}

type NetworkingV1Beta1 struct {
	*networkingv1beta1.Ingress
}

func (n NetworkingV1Beta1) Name() string {
	return n.GetName()
}

func (n NetworkingV1Beta1) IngressClass() (res *string) {
	res = n.Spec.IngressClassName
	if res == nil {
		if a := n.GetAnnotations(); a != nil {
			if v, ok := a[annotationName]; ok {
				res = &v
			}
		}
	}
	return
}

func (n NetworkingV1Beta1) Namespace() string {
	return n.GetNamespace()
}

func (n NetworkingV1Beta1) Hostnames() []string {
	rules := n.Spec.Rules
	var hostnames []string
	for _, rule := range rules {
		hostnames = append(hostnames, rule.Host)
	}
	return hostnames
}

type Extension struct {
	*extensionsv1beta1.Ingress
}

func (e Extension) Name() string {
	return e.GetName()
}

func (e Extension) IngressClass() (res *string) {
	res = e.Spec.IngressClassName
	if res == nil {
		if a := e.GetAnnotations(); a != nil {
			if v, ok := a[annotationName]; ok {
				res = &v
			}
		}
	}
	return
}

func (e Extension) Namespace() string {
	return e.GetNamespace()
}

func (e Extension) Hostnames() []string {
	rules := e.Spec.Rules
	var hostnames []string
	for _, el := range rules {
		hostnames = append(hostnames, el.Host)
	}
	return hostnames
}

type HostnamesList []string

func (hostnames HostnamesList) Len() int {
	return len(hostnames)
}

func (hostnames HostnamesList) Swap(i, j int) {
	hostnames[i], hostnames[j] = hostnames[j], hostnames[i]
}

func (hostnames HostnamesList) Less(i, j int) bool {
	return hostnames[i] < hostnames[j]
}

func (hostnames HostnamesList) IsStringInList(value string) (ok bool) {
	sort.Sort(hostnames)
	i := sort.SearchStrings(hostnames, value)
	ok = i < hostnames.Len() && hostnames[i] == value
	return
}
