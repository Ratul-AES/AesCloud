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

func (r *OrgPodReconciler) reconcileDbPod(ctx context.Context, orgpod *webappv1.OrgPod, l logr.Logger) (webappv1.DbPod, error) {

	dbpod := &webappv1.DbPod{}
	err := r.Get(ctx, types.NamespacedName{Name: orgpod.Name + "-dbpod", Namespace: orgpod.Namespace}, dbpod)
	if err == nil {
		l.Info("DBPOD Found")
		return *dbpod, nil
	}

	if !errors.IsNotFound(err) {
		return *dbpod, err
	}
	l.Info("DEPL Not found, Creating new DEPL")

	labels := map[string]string{
		"dbpod": orgpod.Name + "-dbpod",
	}

	dbpod = &webappv1.DbPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orgpod.Name + "-dbpod",
			Namespace: orgpod.Namespace,
			Labels:    labels,
		},

		Spec: webappv1.DbPodSpec{
			Size:  orgpod.Spec.Size,
			DbImg: orgpod.Spec.DbImg,
		},
	}
	l.Info("Creating Deployment...", "DEPL name", dbpod.Name, "DEPL namespace", dbpod.Namespace)
	if err := ctrl.SetControllerReference(orgpod, dbpod, r.Scheme); err != nil {
		return *dbpod, err
	}

	return *dbpod, r.Create(ctx, dbpod)

}
