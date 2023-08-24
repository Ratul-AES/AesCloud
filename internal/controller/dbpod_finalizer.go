package controller

import (
	"context"

	webappv1 "aes.dev/corepod/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DbPodReconciler) deleteExternalResources(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	l.Info("[DBPOD]: Entered Delete ")
	//Delete Deployment---------> error NOT BEING ABLE TO FIND DEPLOYMENT
	mysql := &appsv1.StatefulSet{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-mysql", Namespace: dbpod.Namespace}, mysql)

	if err != nil {
		return err
	}

	r.Client.Update(ctx, mysql, &client.UpdateOptions{})
	l.Info("[DEPL]: Delete ")

	// Delete Service
	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-svc", Namespace: dbpod.Namespace}, svc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, svc, &client.UpdateOptions{})
	l.Info("[SVC]: Delete ")

	//Delete read Service
	readsvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-readsvc", Namespace: dbpod.Namespace}, readsvc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, readsvc, &client.UpdateOptions{})
	l.Info("[Read SVC]: Delete ")

	//Delete Statefulset PVC
	pvc := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: "data-" + dbpod.Name + "-mysql-0", Namespace: dbpod.Namespace}, pvc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, pvc, &client.UpdateOptions{})
	l.Info("[PVC0]: Delete ")

	pvc = &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: "data-" + dbpod.Name + "-mysql-1", Namespace: dbpod.Namespace}, pvc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, pvc, &client.UpdateOptions{})
	l.Info("[PVC1]: Delete ")

	pvc = &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: "data-" + dbpod.Name + "-mysql-2", Namespace: dbpod.Namespace}, pvc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, pvc, &client.UpdateOptions{})
	l.Info("[PVC2]: Delete ")

	// Delete CMAP
	cmap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-cmap", Namespace: dbpod.Namespace}, cmap)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, cmap, &client.UpdateOptions{})

	l.Info("[CMAP]: Delete ")
	// Delete Secret
	sec := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-sec", Namespace: dbpod.Namespace}, sec)
	if err != nil {
		return err
	}
	r.Client.Update(ctx, sec, &client.UpdateOptions{})
	l.Info("[SECRET]: Delete ")
	// Delete Secret
	init_cmap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-init-db", Namespace: dbpod.Namespace}, init_cmap)
	if err != nil {
		return err
	}
	r.Client.Update(ctx, init_cmap, &client.UpdateOptions{})
	l.Info("[INIT_CMAP]: Delete ")

	//Delete Proxy ConfigMap
	proxyCmap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql-configmap", Namespace: dbpod.Namespace}, proxyCmap)
	if err != nil {
		return err
	}
	r.Client.Update(ctx, proxyCmap, &client.UpdateOptions{})
	l.Info("[PROXY ConfigMap]: Delete ")

	//Delete Proxy Service
	proxyservice := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql", Namespace: dbpod.Namespace}, proxyservice)
	if err != nil {
		return err
	}
	r.Client.Update(ctx, proxyservice, &client.UpdateOptions{})
	l.Info("[PROXY Service]: Delete ")

	//Delete Proxy Statefulset
	proxysts := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql", Namespace: dbpod.Namespace}, proxysts)
	if err != nil {
		return err
	}
	r.Client.Update(ctx, proxysts, &client.UpdateOptions{})
	l.Info("[PROXY Statefulset]: Delete ")

	//Delete Proxy Statefulset PVC
	proxyPvc := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: "my-nfs-pvc-" + dbpod.Name + "-proxysql-0", Namespace: dbpod.Namespace}, proxyPvc)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, pvc, &client.UpdateOptions{})
	l.Info("[PVC0]: Delete ")

	return nil
}
