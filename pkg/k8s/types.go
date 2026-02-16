package k8s

// GeneratorConfig contains configuration for K8s manifest generation
type GeneratorConfig struct {
	Namespace      string
	OutputDir      string
	ImageRegistry  string
	ImageTag       string
	DefaultReplicas int32
}

// DefaultGeneratorConfig returns default configuration
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		Namespace:       "default",
		OutputDir:       "k8s",
		ImageRegistry:   "",
		ImageTag:        "latest",
		DefaultReplicas: 1,
	}
}

// ManifestSet contains all generated K8s manifests
type ManifestSet struct {
	Namespace   string
	Deployments []string
	Services    []string
	ConfigMaps  []string
}
