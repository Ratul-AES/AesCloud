/*
Copyright 2023.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "aes.dev/orgpod/api/v1"
)

// DbPodReconciler reconciles a DbPod object
type DbPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.aes.dev,resources=dbpods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=dbpods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=dbpods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DbPod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *DbPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered DB_POD Reconcile", "req", req)

	//Get DB-POD
	dbpod := &webappv1.DbPod{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, dbpod)

	l.Info("Got dbpod", "spec", dbpod.Spec, "status", dbpod.Status)

	if dbpod.Name == "" {
		l.Info("[DBPOD] Rec entered after delete...")
		return ctrl.Result{}, nil
	}

	dbFinalizer := "webapp.aes.dev/finalizer"
	if dbpod.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(dbpod.GetFinalizers(), dbFinalizer) {
			controllerutil.AddFinalizer(dbpod, dbFinalizer)
			if err := r.Update(ctx, dbpod); err != nil {
				return ctrl.Result{}, err
			}
		}
		l.Info("DBPOD Finalizer not called")
	} else {
		// The object is being deleted
		if containsString(dbpod.GetFinalizers(), dbFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, dbpod, l); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(dbpod, dbFinalizer)
			if err := r.Update(ctx, dbpod); err != nil {
				return ctrl.Result{}, err
			}
			l.Info("DBPOD Finalizer used and removed...")
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if dbpod.Name != dbpod.Status.Name {
		dbpod.Status.Name = dbpod.Name
		r.Status().Update(ctx, dbpod)
	}

	if dbpod.Status.Progress != "Ready" { //Initializing statuses
		dbpod.Status.Progress = "Initiating"
		dbpod.Status.Ready = "0/1"
		r.Status().Update(ctx, dbpod)
	}

	/////////////////// Service	-1
	serv, err := r.reconcileService(ctx, dbpod, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Svc Created :)", "SvcName", serv.Name, "SvcNamespace", serv.Namespace)

	/////////////////// Read Service
	servRead, err := r.reconcileServiceRead(ctx, dbpod, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Read Svc Created :)", "SvcName", servRead.Name, "SvcNamespace", servRead.Namespace)

	///// ENV =====>

	/////////////////// Init DB ConfigMap
	init_cmap, err := r.reconcileInitCmap(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod InitConfigMap Created :)", "CMCName", init_cmap.Name, "CMNamespace", init_cmap.Namespace)

	/////////////////// ConfigMap
	cmap, err := r.reconcileCmap(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod ConfigMap Created :)", "CMCName", cmap.Name, "CMNamespace", cmap.Namespace)

	/////////////////// Secret
	dbsec, err := r.reconcileSecret(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Secret Created :)", "SecretName", dbsec.Name, "SecretNamespace", dbsec.Namespace)

	///// ======> DB DEPL FILE:

	// TODO(user): your logic here   -1
	/*pvc, err := r.reconcilePVC(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod PVC Created :)", "PVCName", pvc.Name, "PVCNamespace", pvc.Namespace)*/

	/////////////////// DB DEPL
	mysql, err := r.reconcileDBDepl(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Created :)", "DB NAME", mysql.Name, "DBNamespace", mysql.Namespace) //after this, for some reason next lines are not being reached
	if mysql.Status.AvailableReplicas > 0 {
		dbpod.Status.Progress = "Ready"
		dbpod.Status.Ready = "1/1"
		r.Status().Update(ctx, dbpod)
	}

	/////////////////// Proxy ConfigMap
	proxyCmap, err := r.reconcileProxyCmap(ctx, dbpod, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Proxy ConfigMap Created :)", "Proxy CMCName", proxyCmap.Name, "Proxy CMNamespace", proxyCmap.Namespace)

	/////////////////// Proxy Service
	proxyserv, err := r.reconcileProxyService(ctx, dbpod, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Proxy Svc Created :)", "Proxy SvcName", proxyserv.Name, "Proxy SvcNamespace", proxyserv.Namespace)

	/////////////////// Proxy Statefulset
	proxySts, err := r.reconcileProxyStatefulset(ctx, dbpod, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("dbpod Proxy Statefulset Created :)", "Proxy SvcName", proxySts.Name, "Proxy Statefulset Namespace", proxySts.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DbPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.DbPod{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
