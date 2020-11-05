package service_labels

import (
	"context"
	"github.com/go-logr/logr"
	discoveryv1alpha1 "k8s.io/api/discovery/v1alpha1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EndpointSlicesLabelsReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	VersionMinor int
	VersionMajor int
}

func (r *EndpointSlicesLabelsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.VersionMajor == 1 && r.VersionMinor == 16 {
		return ctrl.NewControllerManagedBy(mgr).
			For(&discoveryv1alpha1.EndpointSlice{}).
			Complete(r)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1beta1.EndpointSlice{}).
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

	if r.VersionMajor == 1 && r.VersionMinor == 16 {
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
	} else {
		eps := &discoveryv1beta1.EndpointSlice{}
		err = r.Client.Get(ctx, request.NamespacedName, eps)
		if err != nil {
			return reconcile.Result{}, err
		}

		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, eps, func() (err error) {
			eps.ObjectMeta.Labels = sync(eps.ObjectMeta.Labels, tenant)
			eps.ObjectMeta.Annotations = sync(eps.ObjectMeta.Annotations, tenant)
			return nil
		})
	}

	return reconcile.Result{}, err
}
