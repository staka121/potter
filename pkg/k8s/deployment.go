package k8s

import (
	"fmt"
	"strings"

	"github.com/staka121/potter/pkg/types"
)

// GenerateDeployment generates a Kubernetes Deployment manifest
func GenerateDeployment(obj types.ObjectRef, config *GeneratorConfig, tsuboName string) string {
	imageName := getImageName(obj.Name, config.ImageRegistry, config.ImageTag)

	// Generate environment variables for dependencies
	envVars := generateEnvVars(obj.Dependencies, config.Namespace)

	// Generate liveness/readiness probes from health_check
	probes := generateProbes(obj.Runtime.HealthCheck, obj.Runtime.Port)

	manifest := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
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
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
        app.kubernetes.io/name: %s
        app.kubernetes.io/instance: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
          name: http
          protocol: TCP
%s%s
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
`,
		obj.Name,
		config.Namespace,
		obj.Name,
		obj.Name,
		obj.Name,
		tsuboName,
		config.DefaultReplicas,
		obj.Name,
		obj.Name,
		obj.Name,
		obj.Name,
		obj.Name,
		imageName,
		obj.Runtime.Port,
		probes,
		envVars,
	)

	return manifest
}

// getImageName constructs the full image name
func getImageName(serviceName, registry, tag string) string {
	if registry == "" {
		return fmt.Sprintf("%s:%s", serviceName, tag)
	}
	return fmt.Sprintf("%s/%s:%s", registry, serviceName, tag)
}

// generateEnvVars generates environment variables for service dependencies
func generateEnvVars(dependencies []string, namespace string) string {
	if len(dependencies) == 0 {
		return ""
	}

	var envVars strings.Builder
	envVars.WriteString("        env:\n")

	for _, dep := range dependencies {
		// Generate environment variable for each dependency
		// Format: SERVICE_NAME_URL=http://service-name.namespace.svc.cluster.local:port
		envVarName := strings.ToUpper(strings.ReplaceAll(dep, "-", "_")) + "_URL"
		serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", dep, namespace)

		envVars.WriteString(fmt.Sprintf("        - name: %s\n", envVarName))
		envVars.WriteString(fmt.Sprintf("          value: \"%s\"\n", serviceURL))
	}

	return envVars.String()
}

// generateProbes generates liveness and readiness probes
func generateProbes(healthCheckPath string, port int) string {
	if healthCheckPath == "" {
		return ""
	}

	return fmt.Sprintf(`        livenessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
`,
		healthCheckPath, port,
		healthCheckPath, port,
	)
}
