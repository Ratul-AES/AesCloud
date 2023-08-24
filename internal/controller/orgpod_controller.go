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

	webappv1 "aes.dev/corepod/api/v1"
)

// OrgPodReconciler reconciles a OrgPod object
type OrgPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.aes.dev,resources=orgpods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=orgpods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=orgpods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OrgPod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *OrgPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Entered Reconcile", "req", req)

	//Get OrgPod
	org := &webappv1.OrgPod{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, org)

	l.Info("Got Org", "spec", org.Spec, "status", org.Status)
	if org.Name == "" {
		l.Info("[ORGPOD] Rec entered after delete...")
		return ctrl.Result{}, nil
	}
	dbFinalizer := "webapp.aes.dev/finalizerorg"
	if org.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(org.GetFinalizers(), dbFinalizer) {
			controllerutil.AddFinalizer(org, dbFinalizer)
			if err := r.Update(ctx, org); err != nil {
				return ctrl.Result{}, err
			}
		}
		l.Info("ORGPOD Finalizer not called")
	} else {
		// The object is being deleted
		if containsString(org.GetFinalizers(), dbFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, org, l); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(org, dbFinalizer)
			if err := r.Update(ctx, org); err != nil {
				return ctrl.Result{}, err
			}
			l.Info("ORG Finalizer used and removed...")
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if org.Name != org.Status.Name {
		org.Status.Name = org.Name
		r.Status().Update(ctx, org)
	}

	if org.Status.Progress != "Ready" { //Initializing statuses
		org.Status.Progress = "Initiating"
		org.Status.Ready = "0/2"
		r.Status().Update(ctx, org)
	}

	// DB-POD
	dbpod, err := r.reconcileDbPod(ctx, org, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("DB-POD Created", "Name", dbpod.Name, "Namespace", dbpod.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrgPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.OrgPod{}).
		Owns(&webappv1.DbPod{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
