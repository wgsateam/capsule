package service_labels

import (
	"context"
	discoveryv1alpha1 "k8s.io/api/discovery/v1alpha1"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EndpointSlicesLabelsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *EndpointSlicesLabelsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Endpoints{}).
		Complete(r)
}

func (r EndpointSlicesLabelsReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	tenant, err := getTenant(ctx, request.NamespacedName, r.Client)
	if err != nil {
		switch err.(type) {
		case *NonTenantObject, *NoServicesMetadata:
			r.Log.Info(err.Error())
			return reconcile.Result{}, nil
		default:
			r.Log.Error(err, "Cannot sync EndpointSlices labels")
			return reconcile.Result{}, err
		}
	}

	eps := &discoveryv1alpha1.EndpointSlice{}
	err = r.Client.Get(ctx, request.NamespacedName, eps)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, eps, func() (err error) {
		eps.ObjectMeta.Labels = sync(eps.ObjectMeta.Labels, tenant)
		eps.ObjectMeta.Annotations = sync(eps.ObjectMeta.Annotations, tenant)
		return nil
	})

	return reconcile.Result{}, err
}
