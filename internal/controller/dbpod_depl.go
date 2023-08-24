package controller

import (
	"context"

	webappv1 "aes.dev/orgpod/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *DbPodReconciler) reconcileDBDepl(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (appsv1.StatefulSet, error) {

	mysql := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-mysql", Namespace: dbpod.Namespace}, mysql)
	if err == nil {
		l.Info("DEPL Found")
		return *mysql, nil
	}

	if !errors.IsNotFound(err) {
		return *mysql, err
	}
	l.Info("DEPL Not found, Creating new DEPL")

	mysql = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-mysql",
			Namespace: dbpod.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                    dbpod.Name + "-dbapp",
					"app.kubernetes.io/name": "mysql",
				},
			},
			ServiceName: dbpod.Name + "-svc",
			Replicas:    int32Ptr(3),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                    dbpod.Name + "-dbapp",
						"app.kubernetes.io/name": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "init-mysql",
							Image: "mysql:5.7",
							Command: []string{
								"bash",
								"-c",
								`
								set -ex
								# Generate mysql server-id from pod ordinal index.
								[[ $HOSTNAME =~ -([0-9]+)$ ]] || exit 1
								ordinal=${BASH_REMATCH[1]}
								echo [mysqld] > /mnt/conf.d/server-id.cnf
								# Add an offset to avoid reserved server-id=0 value.
								echo server-id=$((100 + $ordinal)) >> /mnt/conf.d/server-id.cnf
								# Copy appropriate conf.d files from config-map to emptyDir.
								if [[ $ordinal -eq 0 ]]; then
									cp /mnt/config-map/primary.cnf /mnt/conf.d/
								else
									cp /mnt/config-map/replica.cnf /mnt/conf.d/
								fi
								`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "conf",
									MountPath: "/mnt/conf.d",
								},
								{
									Name:      "config-map",
									MountPath: "/mnt/config-map",
								},
							},
						},
						{
							Name:  "clone-mysql",
							Image: "gcr.io/google-samples/xtrabackup:1.0",
							Command: []string{
								"bash",
								"-c",
								`
								set -ex
								# Skip the clone if data already exists.
								[[ -d /var/lib/mysql/mysql ]] && exit 0
								# Skip the clone on primary (ordinal index 0).
								[[ $(hostname) =~ -([0-9]+)$ ]] || exit 1
								ordinal=${BASH_REMATCH[1]}
								[[ $ordinal -eq 0 ]] && exit 0
								# Clone data from previous peer.
								ncat --recv-only ` + dbpod.Name + `-mysql-$(($ordinal-1)).` + dbpod.Name + `-svc 3307 | xbstream -x -C /var/lib/mysql
								# Prepare the backup.
								xtrabackup --prepare --target-dir=/var/lib/mysql
								`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql",
									SubPath:   "mysql",
								},
								{
									Name:      "conf",
									MountPath: "/etc/mysql/conf.d",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: "mysql:5.7",
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
									Value: "1",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mysql",
									ContainerPort: 3306,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql",
									SubPath:   "mysql",
								},
								{
									Name:      "conf",
									MountPath: "/etc/mysql/conf.d",
								},
								{
									Name:      "mysql-initdb",
									MountPath: "/docker-entrypoint-initdb.d",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysqladmin",
											"ping",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysql",
											"-h",
											"127.0.0.1",
											"-e",
											"SELECT 1",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
							},
						},
						{
							Name:  "xtrabackup",
							Image: "gcr.io/google-samples/xtrabackup:1.0",
							Ports: []corev1.ContainerPort{
								{
									Name:          "xtrabackup",
									ContainerPort: 3307,
								},
							},
							Command: []string{
								"bash",
								"-c",
								`set -ex
								cd /var/lib/mysql

								# Determine binlog position of cloned data, if any.
								if [[ -f xtrabackup_slave_info && "x$(<xtrabackup_slave_info)" != "x" ]]; then
									# XtraBackup already generated a partial "CHANGE MASTER TO" query
									# because we're cloning from an existing replica. (Need to remove the tailing semicolon!)
									cat xtrabackup_slave_info | sed -E 's/;$//g' > change_master_to.sql.in
									# Ignore xtrabackup_binlog_info in this case (it's useless).
									rm -f xtrabackup_slave_info xtrabackup_binlog_info
								elif [[ -f xtrabackup_binlog_info ]]; then
									# We're cloning directly from primary. Parse binlog position.
									[[ $(cat xtrabackup_binlog_info) =~ ^(.*?)[[:space:]]+(.*?)$ ]] || exit 1
									rm -f xtrabackup_binlog_info xtrabackup_slave_info
									echo "CHANGE MASTER TO MASTER_LOG_FILE='${BASH_REMATCH[1]}',\
										MASTER_LOG_POS=${BASH_REMATCH[2]}" > change_master_to.sql.in
								fi

								# Check if we need to complete a clone by starting replication.
								if [[ -f change_master_to.sql.in ]]; then
									until mysql -h 127.0.0.1 -e "SELECT 1"; do sleep 1; done

									mysql -h 127.0.0.1 -e "$(cat change_master_to.sql.in), \
										MASTER_HOST='` + dbpod.Name + `-mysql-0.` + dbpod.Name + `-svc', \
										MASTER_USER='root', \
										MASTER_PASSWORD='', \
										MASTER_CONNECT_RETRY=10; \
										START SLAVE;"

									mv change_master_to.sql.in change_master_to.sql.orig
								fi

								exec ncat --listen --keep-open --send-only --max-conns=1 3307 -c \
									"xtrabackup --backup --slave-info --stream=xbstream --host=127.0.0.1 --user=root"`,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql",
									SubPath:   "mysql",
								},
								{
									Name:      "conf",
									MountPath: "/etc/mysql/conf.d",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("100Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "conf",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config-map",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: dbpod.Name + "-cmap",
									},
								},
							},
						},
						{
							Name: "mysql-initdb",
							VolumeSource: corev1.VolumeSource{

								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: dbpod.Name + "-init-db",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
						StorageClassName: stringPtr("nfs-client"),
					},
				},
			},
		},
	}

	l.Info("Creating Deployment...", "DEPL name", mysql.Name, "DEPL namespace", mysql.Namespace)
	if err := ctrl.SetControllerReference(dbpod, mysql, r.Scheme); err != nil {
		return *mysql, err
	}

	return *mysql, r.Create(ctx, mysql)
}

func (r *DbPodReconciler) reconcileProxyStatefulset(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (appsv1.StatefulSet, error) {

	proxysts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql", Namespace: dbpod.Namespace}, proxysts)
	if err == nil {
		l.Info("Proxy Statefulset Found")
		return *proxysts, nil
	}

	if !errors.IsNotFound(err) {
		return *proxysts, err
	}

	l.Info("Proxy Statefulset Not found, Creating new Proxy Statefulset")

	proxysts = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-proxysql",
			Namespace: dbpod.Namespace,
			Labels: map[string]string{
				"app": dbpod.Name + "-proxysql",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    int32Ptr(1),
			ServiceName: "proxysqlcluster",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dbpod.Name + "-proxysql",
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dbpod.Name + "-proxysql",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:  "proxysql",
							Image: "proxysql/proxysql:2.3.1",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "proxysql-config",
									MountPath: "/etc/proxysql.cnf",
									SubPath:   "proxysql.cnf",
								},
								{
									Name:      "my-nfs-pvc",
									MountPath: "/var/lib/proxysql",
									SubPath:   "data",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6033,
									Name:          "my-nfs-pvc",
								},
								{
									ContainerPort: 6032,
									Name:          "proxysql-admin",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "proxysql-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: dbpod.Name + "-proxysql-configmap",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-nfs-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: stringPtr("nfs-client"),
					},
				},
			},
		},
	}
	l.Info("Creating Proxy Statefulset...", "Proxy Statefulset name", proxysts.Name, "Proxy Statefulset namespace", proxysts.Namespace)
	if err := ctrl.SetControllerReference(dbpod, proxysts, r.Scheme); err != nil {
		return *proxysts, err
	}

	return *proxysts, r.Create(ctx, proxysts)
}

func int32Ptr(i int32) *int32 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
