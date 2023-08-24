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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "aes.dev/corepod/api/v1"
)

// CorePodReconciler reconciles a CorePod object
type CorePodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.aes.dev,resources=corepods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=corepods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.aes.dev,resources=corepods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CorePod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *CorePodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Entered Reconcile", "req", req)

	//Get OrgPod
	core := &webappv1.CorePod{}
	r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, core)

	l.Info("Got Core", "spec", core.Spec, "status", core.Status)
	if core.Name == "" {
		l.Info("[COREPOD] Rec entered after delete...")
		return ctrl.Result{}, nil
	}
	/*dbFinalizer := "webapp.aes.dev/finalizercore"
	if core.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(core.GetFinalizers(), dbFinalizer) {
			controllerutil.AddFinalizer(core, dbFinalizer)
			if err := r.Update(ctx, core); err != nil {
				return ctrl.Result{}, err
			}
		}
		l.Info("COREPOD Finalizer not called")
	} else {
		// The object is being deleted
		if containsString(core.GetFinalizers(), dbFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, core, l); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(core, dbFinalizer)
			if err := r.Update(ctx, core); err != nil {
				return ctrl.Result{}, err
			}
			l.Info("COREPOD Finalizer used and removed...")
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
	*/
	if core.Name != core.Status.Name {
		core.Status.Name = core.Name
		r.Status().Update(ctx, core)
	}

	if core.Status.Progress != "Ready" { //Initializing statuses
		core.Status.Progress = "Initiating"
		core.Status.Ready = "0/2"
		r.Status().Update(ctx, core)
	}

	//Permissions Resources
	c_ext, err := r.reconcileCoreExt(ctx, core, l)

	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("CoreEXT Created", "CoreEXT Name", c_ext.Name, "CoreEXT Namespace", c_ext.Namespace)

	//Orgpod For BE + DB

	corg, err := r.reconcileCoreOrg(ctx, core, l)

	if err != nil {
		return ctrl.Result{}, err
	}
	l.Info("CoreOrg Created", "CoreOrg Name", corg.Name, "CoreOrg Namespace", corg.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CorePodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.CorePod{}).
		Owns(&webappv1.CoreExt{}).
		Owns(&webappv1.OrgPod{}).
		Complete(r)
}
