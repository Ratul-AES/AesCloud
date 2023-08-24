package controller

import (
	"context"

	webappv1 "aes.dev/orgpod/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *CorePodReconciler) reconcileCoreOrg(ctx context.Context, orgpod *webappv1.CorePod, l logr.Logger) (webappv1.OrgPod, error) {

	corg := &webappv1.OrgPod{}
	err := r.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-org", Namespace: orgpod.Namespace}, corg)
	if err == nil {
		l.Info("DBPOD Found")
		return *corg, nil
	}

	if !errors.IsNotFound(err) {
		return *corg, err
	}
	l.Info("DEPL Not found, Creating new DEPL")

	labels := map[string]string{
		"apps": orgpod.Name + "-org",
	}

	corg = &webappv1.OrgPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orgpod.Name + "-org",
			Namespace: orgpod.Namespace,
			Labels:    labels,
		},

		Spec: webappv1.OrgPodSpec{
			Size:       orgpod.Spec.Size,
			PvSize:     orgpod.Spec.PvSize,
			DbImg:      orgpod.Spec.DbImg,
			BeReplicas: orgpod.Spec.BeReplicas,
			OrgImg:     orgpod.Spec.OrgImg,
		},
	}
	l.Info("Creating Deployment...", "DEPL name", corg.Name, "DEPL namespace", corg.Namespace)
	if err := ctrl.SetControllerReference(orgpod, corg, r.Scheme); err != nil {
		return *corg, err
	}

	return *corg, r.Create(ctx, corg)

}
