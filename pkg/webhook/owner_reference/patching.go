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

package owner_reference

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/clastix/capsule/api/v1alpha1"
	"github.com/clastix/capsule/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-v1-namespace-owner-reference,mutating=true,failurePolicy=fail,groups="",resources=namespaces,verbs=create,versions=v1,name=owner.namespace.capsule.clastix.io

type Webhook struct {
	ForceTenantPrefix bool
}

func (o Webhook) GetHandler() webhook.Handler {
	return &handler{forceTenantPrefix: o.ForceTenantPrefix}
}

func (o Webhook) GetName() string {
	return "OwnerReference"
}

func (o Webhook) GetPath() string {
	return "/mutate-v1-namespace-owner-reference"
}

type handler struct {
	forceTenantPrefix bool
}

func (r *handler) OnCreate(ctx context.Context, req admission.Request, clt client.Client, decoder *admission.Decoder) admission.Response {
	ns := &corev1.Namespace{}
	if err := decoder.Decode(req, ns); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if len(ns.ObjectMeta.Labels) > 0 {
		ln, err := v1alpha1.GetTypeLabel(&v1alpha1.Tenant{})
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		l, ok := ns.ObjectMeta.Labels[ln]
		// assigning namespace to Tenant in case of label
		if ok {
			// retrieving the selected Tenant
			t := &v1alpha1.Tenant{}
			if err := clt.Get(ctx, types.NamespacedName{Name: l}, t); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			// Tenant owner must adhere to user that asked for NS creation
			if !r.isTenantOwner(t.Spec.Owner, req) {
				return admission.Denied("Cannot assign the desired namespace to a non-owned Tenant")
			}
			// Patching the response
			return r.patchResponseForOwnerRef(t, ns)
		}

	}

	// assigning namespace to Tenant in case of --force-tenant-prefix flag enabled
	if r.forceTenantPrefix {
		tenantName := strings.Split(ns.GetName(), "-")[0]
		// retrieving the selected Tenant
		t := &v1alpha1.Tenant{}
		if err := clt.Get(ctx, types.NamespacedName{Name: tenantName}, t); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		// Tenant owner must adhere to user that asked for NS creation
		if !r.isTenantOwner(t.Spec.Owner, req) {
			return admission.Denied("Cannot assign the desired namespace to a non-owned Tenant")
		}
		// Patching the response
		return r.patchResponseForOwnerRef(t, ns)
	}

	tl, err := r.listTenantsForOwner(ctx, "User", req.UserInfo.Username, clt)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if len(tl.Items) > 0 {
		return r.patchResponseForOwnerRef(&tl.Items[0], ns)
	}

	if len(req.UserInfo.Groups) > 0 {
		for _, group := range req.UserInfo.Groups {
			tl, err := r.listTenantsForOwner(ctx, "Group", group, clt)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			if len(tl.Items) > 0 {
				return r.patchResponseForOwnerRef(&tl.Items[0], ns)
			}
		}
	}

	return admission.Denied("You do not have any Tenant assigned: please, reach out the system administrators")
}

func (r *handler) OnDelete(ctx context.Context, req admission.Request, client client.Client, decoder *admission.Decoder) admission.Response {
	return admission.Allowed("")
}

func (r *handler) OnUpdate(ctx context.Context, req admission.Request, client client.Client, decoder *admission.Decoder) admission.Response {
	return admission.Denied("Capsule user cannot update a Namespace")
}

func (r *handler) listTenantsForOwner(ctx context.Context, ownerKind string, ownerName string, clt client.Client) (*v1alpha1.TenantList, error) {
	tl := &v1alpha1.TenantList{}
	f := client.MatchingFields{
		".spec.owner.ownerkind": fmt.Sprintf("%s:%s", ownerKind, ownerName),
	}
	err := clt.List(ctx, tl, f)
	return tl, err
}

func (r *handler) patchResponseForOwnerRef(tenant *v1alpha1.Tenant, ns *corev1.Namespace) admission.Response {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	o, _ := json.Marshal(ns.DeepCopy())
	if err := controllerutil.SetControllerReference(tenant, ns, scheme); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	c, _ := json.Marshal(ns)
	return admission.PatchResponseFromRaw(o, c)
}

func (r *handler) isTenantOwner(os v1alpha1.OwnerSpec, req admission.Request) bool {
	if os.Kind == "User" && req.UserInfo.Username == os.Name {
		return true
	}
	if os.Kind == "Group" {
		for _, group := range req.UserInfo.Groups {
			if group == os.Name {
				return true
			}
		}
	}
	return false
}
