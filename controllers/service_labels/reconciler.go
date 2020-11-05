package service_labels

import (
	"context"
	"github.com/clastix/capsule/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTenant(ctx context.Context, namespacedName types.NamespacedName, client client.Client) (*v1alpha1.Tenant, error) {
	ns := &corev1.Namespace{}
	tenant := &v1alpha1.Tenant{}

	if err := client.Get(ctx, types.NamespacedName{Name: namespacedName.Namespace}, ns); err != nil {
		return nil, err
	}

	capsuleLabel, _ := v1alpha1.GetTypeLabel(&v1alpha1.Tenant{})
	if _, ok := ns.GetLabels()[capsuleLabel]; !ok {
		return nil, NewNonTenantObject(namespacedName.Name)
	}

	if err := client.Get(ctx, types.NamespacedName{Name: ns.Labels[capsuleLabel]}, tenant); err != nil {
		return nil, err
	}

	if tenant.Spec.ServicesMetadata.AdditionalLabels == nil && tenant.Spec.ServicesMetadata.AdditionalAnnotations == nil {
		return nil, NewNoServicesMetadata(namespacedName.Name)
	}

	return tenant, nil
}

func sync(available map[string]string, tenant *v1alpha1.Tenant) map[string]string {
	if al := tenant.Spec.ServicesMetadata.AdditionalLabels; al != nil {
		if available == nil {
			available = make(map[string]string)
			available = tenant.Spec.ServicesMetadata.AdditionalLabels
		} else {
			for key, value := range al {
				if available[key] != value {
					available[key] = value
				}
			}
		}
	}
	return available
}
