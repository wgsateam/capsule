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

package tenantName

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/clastix/capsule/api/v1alpha1"
	"github.com/clastix/capsule/pkg/webhook"
)

// +kubebuilder:webhook:path=/validating-v1-tenant-name,mutating=false,failurePolicy=fail,groups="capsule.clastix.io",resources=tenants,verbs=create,versions=v1alpha1,name=tenant.name.capsule.clastix.io

type Webhook struct{}

func (o Webhook) GetHandler() webhook.Handler {
	return &handler{}
}

func (o Webhook) GetName() string {
	return "TenantName"
}

func (o Webhook) GetPath() string {
	return "/validating-v1-tenant-name"
}

type handler struct{}

func (r *handler) OnCreate(ctx context.Context, req admission.Request, clt client.Client, decoder *admission.Decoder) admission.Response {
	tnt := &v1alpha1.Tenant{}
	if err := decoder.Decode(req, tnt); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	fmt.Printf("%s", tnt.GetName())

	matched, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9]*[a-z0-9])?$`, tnt.GetName())
	if !matched {
		fmt.Printf("%s not mached", tnt.GetName())
		return admission.Denied("Tenant name had forbidden characters")
	}
	return admission.Allowed("")
}

func (r *handler) OnDelete(ctx context.Context, req admission.Request, client client.Client, decoder *admission.Decoder) admission.Response {
	return admission.Allowed("")
}

func (r *handler) OnUpdate(ctx context.Context, req admission.Request, client client.Client, decoder *admission.Decoder) admission.Response {
	return admission.Allowed("")
}
