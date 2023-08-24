package controller

import (
	"context"

	webappv1 "aes.dev/corepod/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *CorePodReconciler) deleteExternalResources(ctx context.Context, corepod *webappv1.CorePod, l logr.Logger) error {
	l.Info("[COREPOD]: Entered Delete ")

	// Delete ORGPOD
	/*orgpod := &webappv1.OrgPod{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: corepod.Name + "-org", Namespace: corepod.Namespace}, orgpod)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, orgpod, &client.UpdateOptions{})
	l.Info("[CORE-ORG]: Delete ")*/

	//Delete Deployment---------> error NOT BEING ABLE TO FIND DEPLOYMENT
	/*gopod := &appsv1.Deployment{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: corepod.Name + "-gopod", Namespace: corepod.Namespace}, gopod)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, gopod, &client.UpdateOptions{})
	l.Info("[GO-RESTAPI]: Delete ")*/

	// Delete Service
	/*svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: corepod.Name + "-goport", Namespace: corepod.Namespace}, svc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, svc, &client.UpdateOptions{})
	l.Info("[GO-PORT]: Delete ")*/

	// Delete CORE_EXT
	corext := &webappv1.CoreExt{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: corepod.Name + "-ext", Namespace: corepod.Namespace}, corext)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, corext, &client.UpdateOptions{})

	return nil
}
