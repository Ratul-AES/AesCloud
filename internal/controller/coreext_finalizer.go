package controller

import (
	"context"

	webappv1 "aes.dev/corepod/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *CoreExtReconciler) deleteExternalResources(ctx context.Context, coreext *webappv1.CoreExt, l logr.Logger) error {
	l.Info("[COREPOD]: Entered Delete ")
	// Delete ROLE BINDING
	rb := &v1.RoleBinding{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: coreext.Name + "-rb", Namespace: coreext.Namespace}, rb)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, rb, &client.UpdateOptions{})
	l.Info("[CORE-RB]: Delete ")

	// Delete CLUSTER ROLE BINDING
	crb := &v1.ClusterRoleBinding{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: coreext.Name + "-crb", Namespace: coreext.Namespace}, crb)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, crb, &client.UpdateOptions{})
	l.Info("[CORE-CRB]: Delete ")

	// Delete ROLE
	role := &v1.Role{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: coreext.Name + "-role", Namespace: coreext.Namespace}, role)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, role, &client.UpdateOptions{})
	l.Info("[CORE-ROLE]: Delete ")

	// Delete CLUSTER ROLE
	crole := &v1.ClusterRole{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: coreext.Name + "-crole", Namespace: coreext.Namespace}, crole)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, crole, &client.UpdateOptions{})
	l.Info("[CORE-CROLE]: Delete ")

	// Delete SERVICE ACCOUNT
	sa := &corev1.ServiceAccount{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: coreext.Name + "-sa", Namespace: coreext.Namespace}, sa)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, sa, &client.UpdateOptions{})
	l.Info("[CORE-SA]: Delete ")

	return nil
}
