package controller

import (
	"context"

	webappv1 "aes.dev/orgpod/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *OrgPodReconciler) deleteExternalResources(ctx context.Context, orgpod *webappv1.OrgPod, l logr.Logger) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	l.Info("[ORGPOD]: Entered Delete ")

	//Delete Deployment---------> error NOT BEING ABLE TO FIND DEPLOYMENT
	orgdelp := &appsv1.Deployment{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-pod", Namespace: orgpod.Namespace}, orgdelp)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, orgdelp, &client.UpdateOptions{})
	l.Info("[DEPL]: Delete ")

	// Delete Service
	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-nodeport", Namespace: orgpod.Namespace}, svc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, svc, &client.UpdateOptions{})
	l.Info("[NODEPORT]: Delete ")
	//Delete PVC
	dbpod := &webappv1.DbPod{}
	err = r.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-dbpod", Namespace: orgpod.Namespace}, dbpod)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, dbpod, &client.UpdateOptions{})
	l.Info("[ORG-DBPOD]: Delete ")

	return nil
}
