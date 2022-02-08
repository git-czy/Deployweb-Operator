/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	webv1 "demo3/api/v1"
)

// DeploywebReconciler reconciles a Deployweb object
type DeploywebReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=web.czy.io,resources=deploywebs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=web.czy.io,resources=deploywebs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=web.czy.io,resources=deploywebs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=replicationcontroller,verbs=get;update;patch;create
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;create

const (
	APPNAME = "web-app"
	PENDING = "pending"
	RUNNING = "running"
	UNKNOWN = "unknown"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
func (r *DeploywebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("starting Reconcile")

	// 查找 deployweb
	deployWeb := &webv1.Deployweb{}
	if err := r.Get(ctx, req.NamespacedName, deployWeb); err != nil {
		if apierrs.IsNotFound(err) {
			logger.Info("unable to fetch deployWeb")
			return reconcile.Result{}, nil
		} else {
			logger.Error(err, "query deployWeb error")
			// 更新状态
			if err := updateStatus(ctx, r, deployWeb, UNKNOWN); err != nil {
				logger.Error(err, "update deployWeb failed in main reconcile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	// 查找ReplicationController
	rcInstance := &corev1.ReplicationController{}
	if err := r.Get(ctx, req.NamespacedName, rcInstance); err != nil {
		if apierrs.IsNotFound(err) {
			logger.Info("rcInstance not exists")
			// 更新状态
			if err = updateStatus(ctx, r, deployWeb, PENDING); err != nil {
				logger.Error(err, "update deployWeb failed in main reconcile")
				return ctrl.Result{}, err
			}
			// 创建 service
			if err := createService(ctx, r, deployWeb, req); err != nil {
				logger.Error(err, "create service failed in main reconcile")
				return ctrl.Result{}, err
			}
			// 创建 rcInstance
			if err := createReplicationController(ctx, r, deployWeb); err != nil {
				logger.Error(err, "create rcInstance failed in main reconcile")
				return ctrl.Result{}, err
			}
			// 更新状态
			if err = updateStatus(ctx, r, deployWeb, RUNNING); err != nil {
				logger.Error(err, "update deployWeb failed in main reconcile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "query rcInstance error")
			// 更新状态
			if err = updateStatus(ctx, r, deployWeb, UNKNOWN); err != nil {
				logger.Error(err, "update deployWeb failed in main reconcile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	deployWebReplica := strconv.Itoa(int(*deployWeb.Spec.Replicas))
	rcInstanceReplica := strconv.Itoa(int(*rcInstance.Spec.Replicas))
	logger.Info("deployWebReplica :" + deployWebReplica + "  rcInstanceReplica :" + rcInstanceReplica)

	if *deployWeb.Spec.Replicas != *rcInstance.Spec.Replicas {
		rcInstance.Spec.Replicas = deployWeb.Spec.Replicas
		if err := r.Update(ctx, rcInstance); err != nil {
			logger.Error(err, "update instance error")
			return ctrl.Result{}, err
		}
		logger.Info("update deployWebReplica from " + rcInstanceReplica + " to " + deployWebReplica)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploywebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.Deployweb{}).
		Complete(r)
}

// createService 创建svc
func createService(ctx context.Context, r *DeploywebReconciler, deployWeb *webv1.Deployweb, req ctrl.Request) error {
	logger := log.FromContext(ctx)
	logger.Info("func createService")

	// 查询service
	service := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, service); err == nil {
		logger.Info("service exists")
		return nil
	} else if err != nil && !apierrs.IsNotFound(err) {
		logger.Error(err, "query service failed")
		return err
	}

	// 实例化一个service
	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deployWeb.Namespace,
			Name:      deployWeb.Name,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					NodePort: deployWeb.Spec.Port,
				},
			},
			Selector: map[string]string{
				"app": APPNAME,
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	// 将service与deployWeb资源关联 级联删除
	logger.Info("set reference from service to deployWeb")
	if err := cutil.SetControllerReference(deployWeb, service, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference error")
		return err
	}

	// 创建service
	logger.Info("start create service")
	if err := r.Create(ctx, service); err != nil {
		logger.Error(err, "create service error")
		return err
	}

	logger.Info("create service success")

	return nil
}

// createReplicationController 创建rc
func createReplicationController(ctx context.Context, r *DeploywebReconciler, deployWeb *webv1.Deployweb) error {
	logger := log.FromContext(ctx)
	logger.Info("func createReplicationController")

	// 实例化一个 rc
	rcInstance := &corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deployWeb.Namespace,
			Name:      deployWeb.Name,
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: deployWeb.Spec.Replicas,
			Selector: map[string]string{
				"app": APPNAME,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": APPNAME,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            APPNAME,
							Image:           deployWeb.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: deployWeb.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}

	// 将rcInstance与deployWeb资源关联 级联删除
	logger.Info("set reference from rcInstance to deployWeb")
	if err := cutil.SetControllerReference(deployWeb, rcInstance, r.Scheme); err != nil {
		logger.Error(err, "SetControllerReference error")
		return err
	}

	// 创建rcInstance
	logger.Info("start create rcInstance")
	if err := r.Create(ctx, rcInstance); err != nil {
		logger.Error(err, "create rcInstance error")
		return err
	}

	logger.Info("create rcInstance success")
	return nil
}

// 更新deployWeb状态
func updateStatus(ctx context.Context, r *DeploywebReconciler, deployWeb *webv1.Deployweb, status string) error {
	logger := log.FromContext(ctx)
	logger.Info("update status :" + status)

	deployWeb.Status.Status = status

	if err := r.Status().Update(ctx, deployWeb); err != nil {
		logger.Error(err, "unable to update deployWeb status")
		return err
	}

	return nil
}
