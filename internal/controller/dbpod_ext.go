package controller

import (
	"context"

	webappv1 "aes.dev/corepod/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *DbPodReconciler) reconcileService(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.Service, error) {

	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-svc", Namespace: dbpod.Namespace}, svc)
	if err == nil {
		l.Info("SVC Found")
		return *svc, nil
	}

	if !errors.IsNotFound(err) {
		return *svc, err
	}

	l.Info("SVC Not found, Creating new SVC")

	labels := map[string]string{
		"app":                    dbpod.Name + "-dbapp",
		"app.kubernetes.io/name": "mysql",
	}

	svc = &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-svc",
			Namespace: dbpod.Namespace,
			Labels:    labels,
		},

		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": dbpod.Name + "-dbapp"},

			Ports: []corev1.ServicePort{
				{
					Port: 3306,
					Name: "mysql",
				},
			},
			ClusterIP: "None",
		},
	}
	l.Info("Creating SVC...", "SVC name", svc.Name, "SVC namespace", svc.Namespace)
	if err := ctrl.SetControllerReference(dbpod, svc, r.Scheme); err != nil {
		return *svc, err
	}

	return *svc, r.Create(ctx, svc)
}

func (r *DbPodReconciler) reconcileServiceRead(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.Service, error) {

	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-readsvc", Namespace: dbpod.Namespace}, svc)
	if err == nil {
		l.Info("Read SVC Found")
		return *svc, nil
	}

	if !errors.IsNotFound(err) {
		return *svc, err
	}

	l.Info("Read SVC Not found, Creating new Read SVC")

	labels := map[string]string{
		"app":                    dbpod.Name + "-dbapp",
		"app.kubernetes.io/name": "mysql",
		"readonly":               "true",
	}

	svc = &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-readsvc",
			Namespace: dbpod.Namespace,
			Labels:    labels,
		},

		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": dbpod.Name + "-dbapp"},

			Ports: []corev1.ServicePort{
				{
					Port: 3306,
					Name: "mysql",
				},
			},
		},
	}
	l.Info("Creating Read SVC...", "Read SVC name", svc.Name, "Read SVC namespace", svc.Namespace)
	if err := ctrl.SetControllerReference(dbpod, svc, r.Scheme); err != nil {
		return *svc, err
	}

	return *svc, r.Create(ctx, svc)
}

func (r *DbPodReconciler) reconcileProxyService(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.Service, error) {

	proxyservice := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql", Namespace: dbpod.Namespace}, proxyservice)
	if err == nil {
		l.Info("Read SVC Found")
		return *proxyservice, nil
	}

	if !errors.IsNotFound(err) {
		return *proxyservice, err
	}

	l.Info("Proxy SVC Not found, Creating new Proxy SVC")

	proxyservice = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-proxysql",
			Namespace: dbpod.Namespace,
			Labels: map[string]string{
				"app": dbpod.Name + "-proxysql",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": dbpod.Name + "-proxysql",
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "proxysql-mysql",
					//NodePort:   30033,
					Port:       6033,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(6033),
				},
				{
					Name: "proxysql-admin",
					//NodePort:   30032,
					Port:       6032,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(6032),
				},
			},
		},
	}
	l.Info("Creating Proxy SVC...", "Proxy SVC name", proxyservice.Name, "Proxy SVC namespace", proxyservice.Namespace)
	if err := ctrl.SetControllerReference(dbpod, proxyservice, r.Scheme); err != nil {
		return *proxyservice, err
	}

	return *proxyservice, r.Create(ctx, proxyservice)
}
