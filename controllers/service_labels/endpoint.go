package service_labels

import (
	"context"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EndpointsLabelsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *EndpointsLabelsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Endpoints{}).
		Complete(r)
}

func (r EndpointsLabelsReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	tenant, err := getTenant(ctx, request.NamespacedName, r.Client)
	if err != nil {
		switch err.(type) {
		case *NonTenantObject, *NoServicesMetadata:
			r.Log.Info(err.Error())
			return reconcile.Result{}, nil
		default:
			r.Log.Error(err, "Cannot sync Endpoints labels")
			return reconcile.Result{}, err
		}
	}

	ep := &corev1.Endpoints{}
	err = r.Client.Get(ctx, request.NamespacedName, ep)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, ep, func() (err error) {
		ep.ObjectMeta.Labels = sync(ep.ObjectMeta.Labels, tenant)
		ep.ObjectMeta.Annotations = sync(ep.ObjectMeta.Annotations, tenant)
		return nil
	})

	return reconcile.Result{}, err
}
