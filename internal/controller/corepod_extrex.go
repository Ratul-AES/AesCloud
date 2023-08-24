package controller

import (
	"context"

	webappv1 "aes.dev/corepod/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *CorePodReconciler) reconcileCoreExt(ctx context.Context, orgpod *webappv1.CorePod, l logr.Logger) (webappv1.CoreExt, error) {
	coreext := &webappv1.CoreExt{}
	err := r.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-ext", Namespace: orgpod.Namespace}, coreext)
	if err == nil {
		l.Info("CoreExt Found")
		return *coreext, nil
	}

	if !errors.IsNotFound(err) {
		return *coreext, err
	}
	l.Info("CoreExt Not found, Creating new CoreExt")

	coreext = &webappv1.CoreExt{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orgpod.Name + "-ext",
			Namespace: orgpod.Namespace,
		},
	}
	l.Info("Creating CoreEXT...", "EXT name", coreext.Name, "EXT namespace", coreext.Namespace)
	if err := ctrl.SetControllerReference(orgpod, coreext, r.Scheme); err != nil {
		return *coreext, err
	}
	return *coreext, r.Create(ctx, coreext)
}
