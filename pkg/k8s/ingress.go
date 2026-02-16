package k8s

import (
	"fmt"
	"strings"

	"github.com/staka121/potter/pkg/types"
)

// IngressConfig contains configuration for Ingress generation
type IngressConfig struct {
	Enabled     bool
	Host        string
	TLSEnabled  bool
	TLSSecretName string
	IngressClass string
	Annotations map[string]string
}

// DefaultIngressConfig returns default Ingress configuration
func DefaultIngressConfig() *IngressConfig {
	return &IngressConfig{
		Enabled:      true,
		Host:         "",
		TLSEnabled:   false,
		IngressClass: "nginx",
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
		},
	}
}

// GenerateIngress generates a Kubernetes Ingress manifest
// This replaces the gateway-service functionality with K8s native Ingress
func GenerateIngress(tsuboDef *types.TsuboDefinition, config *GeneratorConfig, ingressConfig *IngressConfig) string {
	if !ingressConfig.Enabled {
		return ""
	}

	tsuboName := tsuboDef.Tsubo.Name
	namespace := config.Namespace

	// Build Ingress rules based on service contracts
	rules := generateIngressRules(tsuboDef, namespace, ingressConfig)

	// Build TLS configuration if enabled
	tlsConfig := ""
	if ingressConfig.TLSEnabled && ingressConfig.Host != "" {
		tlsConfig = fmt.Sprintf(`  tls:
  - hosts:
    - %s
    secretName: %s
`, ingressConfig.Host, ingressConfig.TLSSecretName)
	}

	// Build host configuration
	hostConfig := ""
	if ingressConfig.Host != "" {
		hostConfig = fmt.Sprintf("    - host: %s\n      http:\n", ingressConfig.Host)
	} else {
		hostConfig = "    - http:\n"
	}

	// Build annotations
	annotations := generateAnnotations(ingressConfig.Annotations, ingressConfig.IngressClass)

	manifest := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s-ingress
  namespace: %s
  labels:
    app.kubernetes.io/name: %s-ingress
    app.kubernetes.io/instance: %s
    app.kubernetes.io/part-of: %s
    app.kubernetes.io/managed-by: potter
    app.kubernetes.io/component: gateway
  annotations:
%s
spec:
%s  rules:
%s%s`,
		tsuboName,
		namespace,
		tsuboName,
		tsuboName,
		tsuboName,
		annotations,
		tlsConfig,
		hostConfig,
		rules,
	)

	return manifest
}

// generateIngressRules generates path-based routing rules
func generateIngressRules(tsuboDef *types.TsuboDefinition, namespace string, ingressConfig *IngressConfig) string {
	var paths []string

	// Generate routes based on gateway-service pattern
	// Each service gets routed based on its base path
	for _, obj := range tsuboDef.Objects {
		// Skip gateway-service itself if it exists
		if obj.Name == "gateway-service" {
			continue
		}

		// Infer the path prefix from service name
		// user-service -> /api/v1/users
		// todo-service -> /api/v1/todos
		pathPrefix := inferPathPrefix(obj.Name)

		path := fmt.Sprintf(`        - path: %s(/|$)(.*)
          pathType: ImplementationSpecific
          backend:
            service:
              name: %s
              port:
                number: 80`,
			pathPrefix,
			obj.Name,
		)
		paths = append(paths, path)
	}

	// Combine all paths under a single "paths:" key
	return "        paths:\n" + strings.Join(paths, "\n")
}

// inferPathPrefix infers the API path prefix from service name
// Examples:
//   user-service -> /api/v1/users
//   todo-service -> /api/v1/todos
//   product-service -> /api/v1/products
func inferPathPrefix(serviceName string) string {
	// Remove "-service" suffix if present
	name := strings.TrimSuffix(serviceName, "-service")

	// Pluralize the name (simple heuristic)
	plural := name
	if !strings.HasSuffix(name, "s") {
		plural = name + "s"
	}

	return fmt.Sprintf("/api/v1/%s", plural)
}

// generateAnnotations generates Ingress annotations
func generateAnnotations(customAnnotations map[string]string, ingressClass string) string {
	annotations := make(map[string]string)

	// Default annotations for nginx ingress controller
	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
	annotations["nginx.ingress.kubernetes.io/use-regex"] = "true"

	// Merge with custom annotations (custom takes precedence)
	for k, v := range customAnnotations {
		annotations[k] = v
	}

	// Add ingress class if not using default
	if ingressClass != "" {
		annotations["kubernetes.io/ingress.class"] = ingressClass
	}

	var result strings.Builder
	for key, value := range annotations {
		result.WriteString(fmt.Sprintf("    %s: \"%s\"\n", key, value))
	}

	return result.String()
}
