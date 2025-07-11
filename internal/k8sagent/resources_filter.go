package k8sagent

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FilterValidResources cross verifies the resources that are present in agent-config
// with those that exist in the cluster such that if resources do not exist in the cluster then they will be excluded
// from the otel-col config
func (km *K8sAgent) FilterValidResources(ctx context.Context, logger *zap.SugaredLogger) {
	valid := *km.Cfg // clone original

	// Filter nodes
	var validNodes []string
	for _, nodeName := range km.Cfg.Monitoring.Nodes.SpecificNodes {
		_, err := km.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err == nil {
			validNodes = append(validNodes, nodeName)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  Node not found: %s\n", nodeName)
		} else {
			logger.Warnf("error validating node %q: %w", nodeName, err)
		}
	}
	valid.Monitoring.Nodes.SpecificNodes = validNodes

	// Filter pods
	var validPods []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}
	for _, pod := range km.Cfg.Monitoring.Pods.SpecificPods {
		_, err := km.K8sClient.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err == nil {
			validPods = append(validPods, pod)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  Pod not found: %s/%s\n", pod.Namespace, pod.Name)
		} else {
			logger.Warnf("error validating pod %s/%s: %w", pod.Namespace, pod.Name, err)
		}
	}
	valid.Monitoring.Pods.SpecificPods = validPods

	// Filter deployments
	var validDeployments []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}
	for _, dep := range km.Cfg.Monitoring.NamedResources.Deployments {
		_, err := km.K8sClient.AppsV1().Deployments(dep.Namespace).Get(ctx, dep.Name, metav1.GetOptions{})
		if err == nil {
			validDeployments = append(validDeployments, dep)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  Deployment not found: %s/%s\n", dep.Namespace, dep.Name)
		} else {
			logger.Warnf("error validating deployment %s/%s: %w", dep.Namespace, dep.Name, err)
		}
	}
	valid.Monitoring.NamedResources.Deployments = validDeployments

	// Filter services
	var validServices []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}
	for _, svc := range km.Cfg.Monitoring.NamedResources.Services {
		_, err := km.K8sClient.CoreV1().Services(svc.Namespace).Get(ctx, svc.Name, metav1.GetOptions{})
		if err == nil {
			validServices = append(validServices, svc)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  Service not found: %s/%s\n", svc.Namespace, svc.Name)
		} else {
			logger.Warnf("error validating service %s/%s: %w", svc.Namespace, svc.Name, err)
		}
	}
	valid.Monitoring.NamedResources.Services = validServices

	// Filter configmaps
	var validConfigMaps []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}
	for _, cm := range km.Cfg.Monitoring.NamedResources.ConfigMaps {
		_, err := km.K8sClient.CoreV1().ConfigMaps(cm.Namespace).Get(ctx, cm.Name, metav1.GetOptions{})
		if err == nil {
			validConfigMaps = append(validConfigMaps, cm)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  ConfigMap not found: %s/%s\n", cm.Namespace, cm.Name)
		} else {
			logger.Warnf("error validating configmap %s/%s: %w", cm.Namespace, cm.Name, err)
		}
	}
	valid.Monitoring.NamedResources.ConfigMaps = validConfigMaps

	// Filter secrets
	var validSecrets []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}
	for _, sec := range km.Cfg.Monitoring.NamedResources.Secrets {
		_, err := km.K8sClient.CoreV1().Secrets(sec.Namespace).Get(ctx, sec.Name, metav1.GetOptions{})
		if err == nil {
			validSecrets = append(validSecrets, sec)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  Secret not found: %s/%s\n", sec.Namespace, sec.Name)
		} else {
			logger.Warnf("error validating secret %s/%s: %w", sec.Namespace, sec.Name, err)
		}
	}
	valid.Monitoring.NamedResources.Secrets = validSecrets

	// Filter PVs
	var validPVs []struct {
		Name string `yaml:"name"`
	}
	for _, pv := range km.Cfg.Monitoring.NamedResources.PersistentVolumes {
		_, err := km.K8sClient.CoreV1().PersistentVolumes().Get(ctx, pv.Name, metav1.GetOptions{})
		if err == nil {
			validPVs = append(validPVs, pv)
		} else if errors.IsNotFound(err) {
			logger.Warnf("⚠️  PersistentVolume not found: %s\n", pv.Name)
		} else {
			logger.Warnf("error validating PV %s: %w", pv.Name, err)
		}
	}
	valid.Monitoring.NamedResources.PersistentVolumes = validPVs
	km.Cfg = &valid
}
