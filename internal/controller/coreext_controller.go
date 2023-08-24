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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "aes.dev/corepod/api/v1"
)

// CoreExtReconciler reconciles a CoreExt object
type CoreExtReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.aes.dev,resources=coreexts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=coreexts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=coreexts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CoreExt object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *CoreExtReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered CoreEXT Reconcile", "req", req)

	//Get ext
	ext := &webappv1.CoreExt{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, ext)

	l.Info("Got EXT", "spec", ext.Spec, "status", ext.Status)

	if ext.Name == "" {
		l.Info("[COREEXT] Rec entered after delete...")
		return ctrl.Result{}, nil
	}
	/*dbFinalizer := "webapp.aes.dev/finalizerext"
	if ext.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(ext.GetFinalizers(), dbFinalizer) {
			controllerutil.AddFinalizer(ext, dbFinalizer)
			if err := r.Update(ctx, ext); err != nil {
				return ctrl.Result{}, err
			}
		}
		l.Info("COREPOD Finalizer not called")
	} else {
		// The object is being deleted
		if containsString(ext.GetFinalizers(), dbFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, ext, l); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(ext, dbFinalizer)
			if err := r.Update(ctx, ext); err != nil {
				return ctrl.Result{}, err
			}
			l.Info("COREEXT Finalizer used and removed...")
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
	*/

	//Permissions Resources

	sa, err := r.reconcileSA(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Service Account Created", "SA Name", sa.Name, "SA Namespace", sa.Namespace)

	role, err := r.reconcileRole(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Role Created", "Role Name", role.Name, "Role Namespace", role.Namespace)

	rb, err := r.reconcileRB(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Rolebinding Created", "RB Name", rb.Name, "RB Namespace", rb.Namespace)

	// TODO(user): your logic here
	cluster_role, err := r.reconcileClusterRole(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Cluster Role Created", "Cluster Role Name", cluster_role.Name, "Cluster Role Namespace", cluster_role.Namespace)

	crb, err := r.reconcileCRB(ctx, ext, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("ClusterR Rolebinding Created", "CRB Name", crb.Name, "CRB Namespace", crb.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreExtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CoreExt{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&v1.Role{}).
		Owns(&v1.RoleBinding{}).
		Complete(r)
}
