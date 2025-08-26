package instrumentation

import (
	"fmt"
	"os"
	"time"
)

type InstrumentAnnotiation map[string]interface{}

// KmCrdAnnotation annotation tells deployment to connect to km-instrumentation crd and enabled/disable the instrumentation
func KmCrdAnnotation(osl string, enabled bool) InstrumentAnnotiation {
	ns := os.Getenv("KM_NAMESPACE")
	if ns == "" {
		ns = "km-agent"
	}
	crd := os.Getenv("KM_CRD_NAME")
	if crd == "" {
		crd = "km-agent-instrumentation-crd"
	}
	lang := ""
	switch osl {
	case "Node.Js":
		lang = "nodejs"
	case "Java":
		lang = "java"
	case "Python":
		lang = "python"
	case "Go":
		lang = "go"
	case "dotnet":
		lang = "dotnet"
	}
	return InstrumentAnnotiation{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						// this annotation will tell k8s api to trigger rollout
						"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
						// contains location/scope of instrumentation crd
						fmt.Sprintf("instrumentation.opentelemetry.io/inject-%s", lang): fmt.Sprintf("%s/%s", ns, crd),
						// TODO: target specific containers
						// "instrumentation.opentelemetry.io/container-names": fmt.Sprintf("%t", enabled),
					},
				},
			},
		},
	}
}
