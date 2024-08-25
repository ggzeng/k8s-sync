package client

import (
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
	"k8sync/pkg/logger"
	"os"
	"strings"

	clientcore "k8s.io/client-go/kubernetes"
	clientreset "k8s.io/client-go/rest"
	clientmetrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8s struct {
	Clientset        clientcore.Interface
	MetricsClientSet *clientmetrics.Clientset
	RestConfig       *clientreset.Config
	namesapce        string // current namespace
	outOfCluster     bool   // out of cluster config
}

// New creates a new k8s client
// cluster - used for get kubeconfig. refer getRestConfig
func New(cluster string) *K8s {
	var err error
	k := K8s{}

	k.RestConfig, err = k.getRestConfig(cluster)
	if err != nil {
		logger.Fatalf("get %s cluster config failed: %v", cluster, err)
		return nil
	}
	k.Clientset, err = clientcore.NewForConfig(k.RestConfig)
	if err != nil {
		logger.Fatalf("can not create kubernetes clientset: %v", err)
		return nil
	}

	k.MetricsClientSet, err = clientmetrics.NewForConfig(k.RestConfig)
	if err != nil {
		logger.Fatalf("can not create kubernetes metric clientset: %v", err)
		return nil
	}
	return &k
}

// GetVersion returns the version of the kubernetes cluster that is running
func (k *K8s) GetVersion() (string, error) {
	version, err := k.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func (k *K8s) SetNamespace(namesapce string) {
	k.namesapce = namesapce
}

func (k *K8s) GetNamespace() string {
	if k.namesapce == "" {
		logger.Warn("can not get current namespace, use 'default'")
		k.namesapce = "default"
	}
	return k.namesapce
}

// GetCurNamespace will return the current namespace for the running program
// Checks for the user passed ENV variable POD_NAMESPACE if not available
// pulls the namespace from pod, if not returns ""
func (k *K8s) GetCurNamespace() string {
	var namespace string
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	if k.outOfCluster {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{})
		namespace, _, _ = kubeconfig.Namespace()
		return namespace
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if namespace = strings.TrimSpace(string(data)); len(namespace) > 0 {
			return namespace
		}
	}
	if namespace == "" {
		logger.Warn("can not get current namespace, use 'default'")
		namespace = "default"
	}
	return namespace
}

func (k *K8s) getRestConfig(cluster string) (*clientreset.Config, error) {
	k.outOfCluster = true
	kubeconfigPath := viper.GetString(cluster + ".kube-config")
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		if kubeconfigPath != "" {
			logger.Infof("get %s cluster default config", cluster)
		}
	} else {
		logger.Infof("get %s cluster config", cluster)
	}
	if kubeconfigPath == "" {
		logger.Infof("use %s cluster internal config", cluster)
		k.outOfCluster = false
		return clientreset.InClusterConfig()
	}
	logger.Infof("use %s cluster out config %s", cluster, kubeconfigPath)
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}
