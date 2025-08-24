package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/kloudmate/km-agent/internal/instrumentation"
	"github.com/kloudmate/km-agent/internal/updater"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func performAPMUpdation(ctx context.Context, response *updater.K8sConfigUpdateResponse) error {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Errorf("failed to build kubeconfig from file: %w", err)
	}

	log.Info("loaded cluster info from In-Cluster service account")
	k8sClient, err := kubernetes.NewForConfig(config)
	if !response.K8s.APMEnabled {
		log.Info("APM not enabled")
		return nil
	}

	for _, app := range response.K8s.APMSettings {
		kind := strings.ToUpper(app.Kind)

		log.Info("KIND :", kind)
		annotations := instrumentation.KmCrdAnnotation(app.Language, app.Enabled)
		annotationBytes, err := json.Marshal(annotations)
		if err != nil {
			return fmt.Errorf("Error marshaling patch %s/%s: %v", app.Namespace, app.Deployment, err)
		}
		switch kind {
		case "DAEMONSET":
			log.Info("Performing APM in :", kind)
			_, err := k8sClient.AppsV1().DaemonSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("Error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		case "DEPLOYMENT":
			log.Info("Performing APM in :", kind)
			if err := handleDeploymentPatching(ctx, k8sClient, app, annotationBytes); err != nil {
				return fmt.Errorf("Error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		case "REPLICASET":
			log.Info("Performing APM ", "kind", kind, "namespace", app.Namespace)
			_, err := k8sClient.AppsV1().ReplicaSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					log.Infof("replicaset not found for %s\n falling back to check for deployment", app.Deployment)
					if err := handleDeploymentPatching(ctx, k8sClient, app, annotationBytes); err != nil {
						return fmt.Errorf("Error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
					}
				} else {
					return fmt.Errorf("Error applying auto instrumentation on %s/%s: %v", app.Deployment, app.Deployment, err)
				}
			}
		case "STATEFULSET":
			log.Info("Performing APM ", "kind", kind, "namespace", app.Namespace)
			_, err := k8sClient.AppsV1().StatefulSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("Error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		default:
			return fmt.Errorf("Error applying auto instrumentation invalid KIND provided %s\n", app.Kind)
		}
	}

	return nil
}

func handleDeploymentPatching(ctx context.Context, client *kubernetes.Clientset, app updater.APMConfig, annotations []byte) error {
	_, err := client.AppsV1().Deployments(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotations, v1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// kubeconfigPath := os.Getenv("KUBECONFIG")
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	// if err != nil {
	// 	log.Errorf("failed to build kubeconfig from file: %w", err)
	// }
	// log.Info("loaded cluster info from In-Cluster service account")
	// k8sClient, err := kubernetes.NewForConfig(config)

	res := updater.K8sConfigUpdateResponse{
		RestartRequired: false,

		K8s: updater.K8sApmConfig{APMEnabled: true, APMSettings: []updater.APMConfig{
			{
				Namespace:  "bookinfo",
				Deployment: "reviews-v2",
				Kind:       "Replicaset",
				Enabled:    true,
				Language:   "Java",
			},
			{
				Namespace:  "bookinfo",
				Deployment: "reviews-v2",
				Kind:       "deployment",
				Enabled:    true,
				Language:   "Java",
			},
			{
				Namespace:  "bookinfo",
				Deployment: "reviews-v2",
				Kind:       "deployment",
				Enabled:    true,
				Language:   "Java",
			},
		},
		},
	}
	err := performAPMUpdation(context.TODO(), &res)
	log.Info(err)
}
