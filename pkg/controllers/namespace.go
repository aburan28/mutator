package namespace

import (
	"context"
	"fmt"
	"pkg/config"

	"github.com/go-logr/logr"
	istio "istio.io/api/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
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
	Config   config.Config
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) ListIngresses(context context.Context, namespace *corev1.Namespace) (bool, error) {
	ingressAnnotation := "alb.ingress.kubernetes.io/load-balancer-name"
	ingressList := &v1.IngressList{}
	var err error
	if err := r.Client.List(context, ingressList, client.InNamespace(namespace.Name)); err != nil {
		for _, ingress := range ingressList.Items {
			annotations := ingress.GetAnnotations()
			if val, exists := annotations[ingressAnnotation]; exists {
				fmt.Printf("Ingress resource in namespace %s has annotation %s: %s\n", namespace.Name, ingressAnnotation, val)
				return true, nil
			}
		}

	}
	return false, err
}

func (r *NamespaceReconciler) ListServices(context context.Context, namespace *corev1.Namespace) (bool, error) {
	serviceAnnotation := "service.beta.kubernetes.io/aws-load-balancer-name"
	services := &corev1.ServiceList{}
	var err error
	if err := r.Client.List(context, services, client.InNamespace(namespace.Name)); err != nil {
		for _, service := range services.Items {
			annotations := service.GetAnnotations()
			if val, exists := annotations[serviceAnnotation]; exists {
				fmt.Printf("Service resource in namespace %s has annotation %s: %s\n", namespace.Name, serviceAnnotation, val)
				return true, nil
			}
		}
	}
	return false, err
}

func (r *NamespaceReconciler) ListPeerAuthentications(context context.Context, namespace *corev1.Namespace) (bool, error) {
	paList := &istio.PeerAuthenticationList{}
	var err error
	if err := r.Client.List(context, paList, client.InNamespace(namespace.Name)); err != nil {
		if err != nil {
			fmt.Printf("Error listing PeerAuthentication resources: %v\n", err)
			return false, err
		}

		if len(paList.Items) > 0 {
			fmt.Printf("Found %d PeerAuthentication resources in namespace %s\n", len(paList.Items), namespace)
			return true, nil
			// Iterate through paList.Items to access individual resources.
		} else {
			fmt.Printf("No PeerAuthentication resources found in namespace %s\n", namespace)
			// Handle the case where no resources were found.
		}
	}
	return false, err
}

func (r *NamespaceReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	var namespace corev1.Namespace
	if err := r.Client.Get(context, req.NamespacedName, &namespace); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Info("namespace", namespace)
		labels := namespace.Labels
		annotations := namespace.Annotations

		if labels == nil {
			labels = make(map[string]string)
		}

		if annotations == nil {
			annotations = make(map[string]string)
		}

		alb, err := r.ListIngresses(context, &namespace)
		if err != nil {
			r.Log.Error(err, "unable to list Ingresses")
		}
		nlb, err := r.ListServices(context, &namespace)
		if err != nil {
			r.Log.Error(err, "unable to list Services")
		}

		if alb || nlb {
			labels["elbv2.k8s.aws/pod-readiness-gate-inject"] = "enabled"
		}

		pa, err := r.ListPeerAuthentications(context, &namespace)
		if err != nil {
			r.Log.Error(err, "unable to list PeerAuthentications")
		}
		if pa {
			labels["istio-injection"] = "enabled"
		}

		namespace.Labels = labels
		namespace.Annotations = annotations

		r.Log.Error(err, "unable to fetch Namespace")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
