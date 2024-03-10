package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdsv1 "github.com/abdheshnayak/kubewiremesh/api/v1"
	"github.com/abdheshnayak/kubewiremesh/controllers/env"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConnectReconciler reconciles a Connect object
type ConnectReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	logger     logging.Logger
	Name       string
	yamlClient kubectl.YAMLClient
	Env        *env.Env
}

const (
	ConnectNameKey = "anayak.com.np/connect.name"
	ConnectMarkKey = "anayak.com.np/connect"
)

/*
steps to implement:
1. ensure private key, public key and ip are set
2. update configmap from services
3. replicate service of another cluster
4. update service from services, to send request to another cluster
5. update deployment from services, to send request to another cluster
6. update deployment to get request from another cluster
*/

//+kubebuilder:rbac:groups=crds.anayak.com.np,resources=connects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.anayak.com.np,resources=connects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.anayak.com.np,resources=connects/finalizers,verbs=update

func (r *ConnectReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := rApi.NewRequest(rApi.NewReconcilerCtx(ctx, r.logger), r.Client, request.NamespacedName, &crdsv1.Connect{})
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if req.Object.GetDeletionTimestamp() != nil {
		if x := r.finalize(req); !x.ShouldProceed() {
			return x.ReconcilerResponse()
		}

		return ctrl.Result{}, nil
	}

	req.PreReconcile()
	defer req.PostReconcile()

	if step := req.ClearStatusIfAnnotated(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureChecks(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureLabelsAndAnnotations(); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := req.EnsureFinalizers(constants.ForegroundFinalizer, constants.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	// if step := r.reconNodeportConfigAndSvc(req); !step.ShouldProceed() {
	// 	return step.ReconcilerResponse()
	// }

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ConnectReconciler) finalize(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	return req.Finalize()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectReconciler) SetupWithManager(mgr ctrl.Manager, logger logging.Logger) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.logger = logger.WithName(r.Name)
	r.yamlClient = kubectl.NewYAMLClientOrDie(mgr.GetConfig(), kubectl.YAMLClientOpts{Logger: r.logger})

	builder := ctrl.NewControllerManagedBy(mgr)

	builder.For(&crdsv1.Connect{})

	watchlist := []client.Object{
		&corev1.Service{},
		&corev1.ConfigMap{},
		&appsv1.Deployment{},
	}

	for _, obj := range watchlist {
		builder.Watches(obj, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			// result := []reconcile.Request{}
			// if o.GetAnnotations()[SvcMarkKey] != "true" {
			// 	return result
			// }
			//
			// pbsList := &crdsv1.PortBridgeServiceList{}
			// if err := r.List(ctx, pbsList); err != nil {
			// 	return nil
			// }
			//
			// for _, pbs := range pbsList.Items {
			// 	if slices.Contains(pbs.Spec.Namespaces, o.GetNamespace()) || o.GetNamespace() == "default" {
			// 		result = append(result, reconcile.Request{
			// 			NamespacedName: client.ObjectKey{
			// 				Name: pbs.GetName(),
			// 			},
			// 		})
			// 	}
			// }
			//
			// return result

			return []reconcile.Request{}
		}))
	}

	return builder.Complete(r)

}
