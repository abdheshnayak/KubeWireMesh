package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdsv1 "github.com/abdheshnayak/kubewiremesh/api/v1"
	"github.com/abdheshnayak/kubewiremesh/controllers/constants"
	"github.com/abdheshnayak/kubewiremesh/controllers/env"
	"github.com/abdheshnayak/kubewiremesh/controllers/templates"
	"github.com/abdheshnayak/kubewiremesh/controllers/types"
	"github.com/abdheshnayak/kubewiremesh/controllers/utils"
	cn "github.com/kloudlite/operator/pkg/constants"
	apiLabels "k8s.io/apimachinery/pkg/labels"

	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/kubectl"
	"github.com/kloudlite/operator/pkg/logging"
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	ConnectNameKey string = "anayak.com.np/kubewiremesh-connect.name"
	ConnectMarkKey string = "anayak.com.np/kubewiremesh-connect"
	MarkExposedKey string = "anayak.com.np/kubewiremesh-connect.exposed"
)

const (
	KeysAndIpReady string = "KeysAndIpReady"
	ConfigMapReady string = "ConfigMapReady"

	PeersAvailable string = "PeersAvailable"

	ReceiverReady     string = "ReceiverReady"
	ReceiverConfReady string = "ReceiverConfReady"

	VirtualServicesReady string = "VirtualServicesReady"
)

/*
steps to implement:
1. Ensure private key, public key and ip are set
2. Update configmap from services

3. Replicate service of another cluster
4. Update service from services, to send request to another cluster
5. Update deployment from services, to send request to another cluster
6. Update deployment to get request from another cluster
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

	if step := req.EnsureFinalizers(cn.ForegroundFinalizer, cn.CommonFinalizer); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.reconKeysAndIp(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	if step := r.updateConfigMap(req); !step.ShouldProceed() {
		return step.ReconcilerResponse()
	}

	req.Object.Status.IsReady = true
	return ctrl.Result{}, nil
}

func (r *ConnectReconciler) reconKeysAndIp(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(KeysAndIpReady, req)

	updated := false

	if obj.Spec.PrivateKey == nil {
		pub, priv, err := utils.GenerateWgKeys()
		if err != nil {
			return check.Failed(err)
		}

		obj.Spec.PrivateKey = utils.Ptr(string(priv))
		obj.Spec.PublicKey = utils.Ptr(string(pub))

		updated = true
	}

	if obj.Spec.PublicKey == nil {
		pub, err := utils.GeneratePublicKey(*obj.Spec.PrivateKey)
		if err != nil {
			return check.Failed(err)

		}

		obj.Spec.PublicKey = utils.Ptr(string(pub))
		updated = true
	}

	if obj.Spec.Ip == nil {
		ip, err := utils.GetRemoteDeviceIp(int64(obj.Spec.Id))
		if err != nil {
			return check.Failed(err)

		}

		obj.Spec.Ip = utils.Ptr(string(ip))
		updated = true
	}

	if updated {
		if err := r.Update(ctx, obj); err != nil {
			return check.Failed(err)
		}
	}

	return check.Completed()
}

func (r *ConnectReconciler) updateConfigMap(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ConfigMapReady, req)

	var services corev1.ServiceList
	if err := r.List(ctx, &services, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			MarkExposedKey: "true",
		}),
	}); err != nil {
		r.logger.Error(err)
	}

	occupiedPorts := utils.PortMap{}
	if cm, err := rApi.Get(ctx, r.Client, fn.NN("default", fmt.Sprintf("%s-config", obj.Name)), &corev1.ConfigMap{}); err == nil {
		s := cm.Data["occupiedPorts"]
		occupiedPorts.ParseBytes([]byte(s))
	}

	var data utils.PortMap
	for _, svc := range services.Items {
		for _, port := range svc.Spec.Ports {
			pd := utils.PortData{
				Namespace: svc.Namespace,
				Name:      svc.Name,
				Port:      port.Port,
			}

			if data.SvcExist(pd) {
				continue
			}

			if occupiedPorts.SvcExist(pd) {
				data.AddPort(*occupiedPorts.GetPort(pd), pd)
				continue
			}

			data.AddPort(data.GetRandomPort(occupiedPorts), pd)
		}
	}

	if !data.IsEquals(occupiedPorts) {
		bytes, err := data.ToBytes()
		if err != nil {
			return check.Failed(err)
		}

		cm := &corev1.ConfigMap{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("%s-config", obj.Name),
				Namespace: "default",
			},
			Data: map[string]string{
				"occupiedPorts": string(bytes),
			},
		}

		if err := fn.KubectlApply(ctx, r.Client, cm); err != nil {
			return check.Failed(err)
		}

		for _, p := range obj.Spec.Peers {
			ip, err := utils.GetRemoteDeviceIp(int64(p.Id))
			if err != nil {
				r.logger.Errorf(err, "Failed to get remote device ip for peer %d", p.Id)
				continue
			}

			if err := utils.SendBytesToReceiver(ip, bytes); err != nil {
				r.logger.Errorf(err, "Failed to send occupied ports to %s", ip)
				continue
			}
		}
	}

	return check.Completed()
}

func (r *ConnectReconciler) ensurePeersAvailable(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	_, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(PeersAvailable, req)

	if len(obj.Spec.Peers) == 0 {
		return check.Failed(fmt.Errorf("No peers available"))
	}

	return check.Completed()
}

func (r *ConnectReconciler) reconReceiverWgConf(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	ctx, obj := req.Context(), req.Object
	check := rApi.NewRunningCheck(ReceiverConfReady, req)

	templateBuildJob, err := templates.Read(templates.WGConfig)
	if err != nil {
		return check.Failed(err)
	}

	o := types.WireguardConfig{}

	b, err := templates.ParseBytes(templateBuildJob, o)
	if err != nil {
		return check.Failed(err)
	}

	secName := fmt.Sprintf("%s-receiver-conf", obj.Name)

	s, err := rApi.Get(ctx, r.Client, fn.NN(constants.CONN_DATA_NS, secName), &corev1.Secret{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}
	}

	if s != nil {
		b := s.Data["wg0.conf"]
		if string(b) == string(b) {
			return check.Completed()
		}
	}

	if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      secName,
			Namespace: constants.CONN_DATA_NS,
		},
		Data: map[string][]byte{
			"wg0.conf": b,
		},
	}); err != nil {
		return check.Failed(err)
	}

	return nil
}

// create receiver pod
func (r *ConnectReconciler) reconReceiver(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	return nil
}

func (r *ConnectReconciler) reconSender(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	return nil
}

func (r *ConnectReconciler) finalize(req *rApi.Request[*crdsv1.Connect]) stepResult.Result {
	return nil
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
