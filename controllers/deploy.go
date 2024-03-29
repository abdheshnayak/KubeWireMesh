package controllers

import (
	"fmt"

	crdsv1 "github.com/abdheshnayak/kubewiremesh/api/v1"
	"github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getReceiverDeployment(req *rApi.Request[*crdsv1.Connect]) (*appsv1.Deployment, error) {
	obj, name := req.Object, req.Object.Name

	labels := map[string]string{}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-deployment", name),
			Namespace:       "default",
			Labels:          labels,
			Annotations:     labels,
			OwnerReferences: []metav1.OwnerReference{functions.AsOwner(obj, true)},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{
									"NET_ADMIN",
								},
							},
						},
						Command: func() []string {
							res := []string{
								"/bin/sh",
								"-c",
							}

							cmd := "apk add --no-cache iptables\n"

							cmd += "iptables -t nat -A POSTROUTING -j MASQUERADE\n"

							// for nodeport, data := range nodeports {
							// 	if data.Ip == "" {
							// 		continue
							// 	}
							//
							// 	cmd += fmt.Sprintf("iptables -t nat -A OUTPUT -p %s --dport %d -d %s -j DNAT --to-destination %s:%d\n",
							// 		data.Protocol, nodeport, "127.0.0.1", data.Ip, data.Port)
							// }

							cmd += "tail -f /dev/null"

							res = append(res, cmd)
							return res
						}(),
						Name:  "port-bridge",
						Image: "alpine:latest",
					}},
				},
			},
		},
	}, nil
}

func (r *ConnectReconciler) upsertService(req *rApi.Request[*crdsv1.Connect]) error {
	return nil
}
