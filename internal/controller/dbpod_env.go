package controller

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	webappv1 "aes.dev/orgpod/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
)

// DB CMAP
func (r *DbPodReconciler) reconcileCmap(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.ConfigMap, error) {
	cmap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-cmap", Namespace: dbpod.Namespace}, cmap)
	if err == nil {
		l.Info("SVC Found")
		return *cmap, nil
	}

	if !errors.IsNotFound(err) {
		return *cmap, err
	}

	labels := map[string]string{
		"app":                    dbpod.Name + "-dbapp",
		"app.kubernetes.io/name": "mysql",
	}
	cmap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-cmap",
			Namespace: dbpod.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"host":    dbpod.Name + "-proxysql:6033",
			"db_name": "demo_proj",
			"primary.cnf": `# Apply this config only on the primary.
							[mysqld]
							log-bin`,
			"replica.cnf": `# Apply this config only on replicas.
							[mysqld]
							super-read-only`,
		},
	}
	l.Info("Creating OrgNodePort...", "NodePort name", cmap.Name, "NodePort namespace", cmap.Namespace)
	if err := ctrl.SetControllerReference(dbpod, cmap, r.Scheme); err != nil {
		return *cmap, err
	}

	return *cmap, r.Create(ctx, cmap)
}

// DB Secret
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (r *DbPodReconciler) reconcileSecret(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.Secret, error) {
	sec := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-sec", Namespace: dbpod.Namespace}, sec)
	if err == nil {
		l.Info("Secret Found")
		return *sec, nil
	}

	if !errors.IsNotFound(err) {
		return *sec, err
	}

	l.Info("Secret Not found, Creating new secret")
	sec = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-sec",
			Namespace: dbpod.Namespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"passwordRoot": []byte(StringWithCharset(10, charset)),
			"username":     []byte(dbpod.Name),
			"password":     []byte(StringWithCharset(10, charset)),
		},
	}
	l.Info("Creating Secret...", "Secret name", sec.Name, "Secret namespace", sec.Namespace)
	if err := ctrl.SetControllerReference(dbpod, sec, r.Scheme); err != nil {
		return *sec, err
	}

	return *sec, r.Create(ctx, sec)
}

// Config Map to initialize db
func (r *DbPodReconciler) reconcileInitCmap(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.ConfigMap, error) {
	cmap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-init-db", Namespace: dbpod.Namespace}, cmap)
	if err == nil {
		l.Info("Cmap Found")
		return *cmap, nil
	}

	if !errors.IsNotFound(err) {
		return *cmap, err
	}

	secretName := dbpod.Name + "-sec"
	secretNamespace := dbpod.Namespace

	secret, err := GetSecret(secretName, secretNamespace)
	if err != nil {
		fmt.Printf("Error fetching Secret: %v\n", err)
		return *cmap, nil
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])

	// Create the init.sql script
	initSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS demo_proj; USE demo_proj;"+
		"CREATE USER '%s' IDENTIFIED BY '%s';"+
		"GRANT ALL PRIVILEGES ON demo_proj.* TO '%s';", username, password, username)

	labels := map[string]string{
		"app":                    dbpod.Name + "-dbapp",
		"app.kubernetes.io/name": "mysql",
	}
	cmap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-init-db",
			Namespace: dbpod.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"init.sql": initSQL,
		},
	}
	l.Info("Creating DB Init ConfigMap...", "cmap name", cmap.Name, "cmap namespace", cmap.Namespace)
	if err := ctrl.SetControllerReference(dbpod, cmap, r.Scheme); err != nil {
		return *cmap, err
	}

	return *cmap, r.Create(ctx, cmap)
}

// PrxySQl ConfigMap
func (r *DbPodReconciler) reconcileProxyCmap(ctx context.Context, dbpod *webappv1.DbPod, l logr.Logger) (corev1.ConfigMap, error) {
	proxyCmap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: dbpod.Name + "-proxysql-configmap", Namespace: dbpod.Namespace}, proxyCmap)
	if err == nil {
		l.Info("Proxy ConfigMap Found")
		return *proxyCmap, nil
	}

	if !errors.IsNotFound(err) {
		return *proxyCmap, err
	}

	secretName := dbpod.Name + "-sec"
	secretNamespace := dbpod.Namespace

	secret, err := GetSecret(secretName, secretNamespace)
	if err != nil {
		l.Info("Error fetching Secret: %v\n", err)
		return *proxyCmap, nil
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	l.Info("UserName: " + username + " Password: " + password)

	proxyCmap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbpod.Name + "-proxysql-configmap",
			Namespace: dbpod.Namespace,
			Labels: map[string]string{
				"app": dbpod.Name + "-proxysql",
			},
		},
		Data: map[string]string{
			"proxysql.cnf": `
				datadir="/var/lib/proxysql"
	
				admin_variables=
				{
					admin_credentials="admin:admin;cluster:secret"
					mysql_ifaces="0.0.0.0:6032"
					refresh_interval=2000
					cluster_username="cluster"
					cluster_password="secret"  
				}
				
				mysql_variables=
				{
					threads=4
					max_connections=2048
					default_query_delay=0
					default_query_timeout=36000000
					have_compress=true
					poll_timeout=2000
					interfaces="0.0.0.0:6033;/tmp/proxysql.sock"
					default_schema="information_schema"
					stacksize=1048576
					server_version="8.0.23"
					connect_timeout_server=3000
					monitor_username="monitor"
					monitor_password="monitor"
					monitor_history=600000
					monitor_connect_interval=60000
					monitor_ping_interval=10000
					monitor_read_only_interval=1500
					monitor_read_only_timeout=500
					ping_interval_server_msec=120000
					ping_timeout_server=500
					commands_stats=true
					sessions_sort=true
					connect_retries_on_failure=10
				}
			
				mysql_servers =
				(
					{ address="` + dbpod.Name + `-readsvc" , port=3306 , hostgroup=10, max_connections=100 }, #cip name
					{ address="` + dbpod.Name + `-mysql-0.` + dbpod.Name + `-svc" , port=3306 , hostgroup=20, max_connections=100 } #hostfrp 20 = master #headless name
				)
				
				mysql_users =
				(
					{ username = "` + username + `", password = "` + password + `", default_hostgroup = 20, active = 1 }
				)
			
			
			
				mysql_query_rules =
				(
			
			
						# ALL THE DATABASE, TABLE, SCHEMA MANIPULATION COMMANDS----> MASTER(20)
						{
								rule_id=100
								active=1
								match_pattern="^(CREATE|ALTER|DROP|ALTER|TRUNCATE|RENAME|MERGE|INSERT|REPLACE|DELETE|LOAD|CALL|SET) .*$"
								destination_hostgroup=20
								apply=1
						},
			
						# LOCK TABLE FOR UPDATE ----------------------------------> MASTER(20)
						{
								rule_id=200
								active=1
								match_pattern="^SELECT .* FOR UPDATE"
								destination_hostgroup=20
								apply=1
						},
						# ALL OTHER READS   ----------------------------------------> SLAVE(10)
						{
								rule_id=300
								active=1
								match_pattern="^(SELECT|SHOW|DESCRIBE|EXPLAIN) .*"
								destination_hostgroup=10
								apply=1
						},
			
				)		
				`,
		},
	}
	l.Info("Creating Proxy configMap...", "Proxy configMap name", proxyCmap.Name, "Proxy configMap namespace", proxyCmap.Namespace)
	if err := ctrl.SetControllerReference(dbpod, proxyCmap, r.Scheme); err != nil {
		return *proxyCmap, err
	}

	return *proxyCmap, r.Create(ctx, proxyCmap)
}

func GetSecret(secretName, namespace string) (*corev1.Secret, error) {
	// Initialize the Kubernetes client
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = ""
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Fetch the Secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}
