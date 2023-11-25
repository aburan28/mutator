package namespace

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciler struct {
	Client   client.Client
	Log      logr.Logger
	Lister   client.Reader
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	var namespace corev1.Namespace
	if err := r.Client.Get(context, req.NamespacedName, &namespace); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to fetch Namespace")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
