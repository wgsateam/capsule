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

type ServicesLabelsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ServicesLabelsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

func (r ServicesLabelsReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	tenant, err := getTenant(ctx, request.NamespacedName, r.Client)
	if err != nil {
		switch err.(type) {
		case *NonTenantObject, *NoServicesMetadata:
			r.Log.Info(err.Error())
			return reconcile.Result{}, nil
		default:
			r.Log.Error(err, "Cannot sync Service labels")
			return reconcile.Result{}, err
		}
	}

	service := &corev1.Service{}
	err = r.Client.Get(ctx, request.NamespacedName, service)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, service, func() (err error) {
		service.ObjectMeta.Labels = sync(service.ObjectMeta.Labels, tenant)
		service.ObjectMeta.Annotations = sync(service.ObjectMeta.Annotations, tenant)
		return nil
	})

	return reconcile.Result{}, err
}
