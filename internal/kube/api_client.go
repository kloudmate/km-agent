package kube

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func initKubeClient(logger *zap.SugaredLogger) (*kubernetes.Clientset, error) {
	// Try kubeconfig (out-of-cluster)
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from KUBECONFIG: %w", err)
		}
		return kubernetes.NewForConfig(cfg)
	}

	// Fallback to in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}
