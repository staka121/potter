package k8s

import (
	"fmt"
)

// GenerateNamespace generates a Kubernetes Namespace manifest
func GenerateNamespace(namespace, tsuboName string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    app.kubernetes.io/name: %s
    app.kubernetes.io/managed-by: potter
`, namespace, tsuboName)
}
