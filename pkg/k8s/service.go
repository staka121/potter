package k8s

import (
	"fmt"

	"github.com/staka121/potter/pkg/types"
)

// GenerateService generates a Kubernetes Service manifest
func GenerateService(obj types.ObjectRef, config *GeneratorConfig, tsuboName string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
    app.kubernetes.io/name: %s
    app.kubernetes.io/instance: %s
    app.kubernetes.io/part-of: %s
    app.kubernetes.io/managed-by: potter
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: %d
    protocol: TCP
    name: http
  selector:
    app: %s
`,
		obj.Name,
		config.Namespace,
		obj.Name,
		obj.Name,
		obj.Name,
		tsuboName,
		obj.Runtime.Port,
		obj.Name,
	)
}
