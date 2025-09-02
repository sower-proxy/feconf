package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/sower-proxy/conf/reader"
)

// Test function for URI parsing without creating actual Kubernetes client
func TestParseURI(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		wantErr      bool
		resourceType string
		namespace    string
		configName   string
		key          string
	}{
		{
			name:         "valid configmap uri",
			uri:          "k8s://configmap/default/my-config",
			wantErr:      false,
			resourceType: ResourceTypeConfigMap,
			namespace:    "default",
			configName:   "my-config",
			key:          "",
		},
		{
			name:         "valid configmap uri with key",
			uri:          "k8s://configmap/default/my-config/config.yaml",
			wantErr:      false,
			resourceType: ResourceTypeConfigMap,
			namespace:    "default",
			configName:   "my-config",
			key:          "config.yaml",
		},
		{
			name:         "valid secret uri",
			uri:          "k8s://secret/default/my-secret",
			wantErr:      false,
			resourceType: ResourceTypeSecret,
			namespace:    "default",
			configName:   "my-secret",
			key:          "",
		},
		{
			name:         "valid secret uri with key",
			uri:          "k8s://secret/default/my-secret/password",
			wantErr:      false,
			resourceType: ResourceTypeSecret,
			namespace:    "default",
			configName:   "my-secret",
			key:          "password",
		},
		{
			name:    "invalid scheme",
			uri:     "http://configmap/default/my-config",
			wantErr: true,
		},
		{
			name:    "invalid resource type",
			uri:     "k8s://deployment/default/my-deployment",
			wantErr: true,
		},
		{
			name:    "missing parts",
			uri:     "k8s://configmap/default",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse URI manually without creating client
			r, err := parseURIManually(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURIManually() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if r.resourceType != tt.resourceType {
					t.Errorf("parseURIManually() resourceType = %v, want %v", r.resourceType, tt.resourceType)
				}
				if r.namespace != tt.namespace {
					t.Errorf("parseURIManually() namespace = %v, want %v", r.namespace, tt.namespace)
				}
				if r.name != tt.configName {
					t.Errorf("parseURIManually() name = %v, want %v", r.name, tt.configName)
				}
				if r.key != tt.key {
					t.Errorf("parseURIManually() key = %v, want %v", r.key, tt.key)
				}
			}
		})
	}
}

// parseURIManually parses URI without creating Kubernetes client for testing
func parseURIManually(uri string) (*K8SReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeK8S) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s", u.Scheme, SchemeK8S)
	}

	// Parse URI: k8s://{resourceType}/{namespace}/{name}[/{key}]
	// Example: k8s://configmap/default/my-config/config.yaml
	// In this case:
	// - Host (u.Host) is the resourceType (configmap)
	// - Path (u.Path) is /{namespace}/{name}[/{key}]

	resourceType := u.Host

	// Validate resource type
	if resourceType != ResourceTypeConfigMap && resourceType != ResourceTypeSecret {
		return nil, fmt.Errorf("unsupported resource type: %s, expected: %s or %s", resourceType, ResourceTypeConfigMap, ResourceTypeSecret)
	}

	// Parse path: /{namespace}/{name}[/{key}]
	path := u.Path
	// Remove leading slash if present
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Split path properly
	parts := []string{}
	start := 0
	for i, char := range path {
		if char == '/' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	// Add the last part
	if start < len(path) {
		parts = append(parts, path[start:])
	}

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid k8s URI path, expected format: k8s://{resourceType}/{namespace}/{name}[/{key}]")
	}

	namespace := parts[0]
	name := parts[1]
	key := ""
	if len(parts) > 2 {
		key = parts[2]
	}

	return &K8SReader{
		uri:          uri,
		resourceType: resourceType,
		namespace:    namespace,
		name:         name,
		key:          key,
	}, nil
}

func TestK8SReader_Interface(t *testing.T) {
	// This test ensures K8SReader implements ConfReader interface
	var _ reader.ConfReader = &K8SReader{}
}

func TestK8SReader_Read_NotImplemented(t *testing.T) {
	// This test is just a placeholder since we can't easily test actual k8s calls
	// In a real test environment, we would mock the k8s client
	r := &K8SReader{
		resourceType: ResourceTypeConfigMap,
		namespace:    "default",
		name:         "test-config",
		clientset:    nil, // This will cause a specific error when trying to use it
	}
	
	ctx := context.Background()
	_, err := r.Read(ctx)
	// We expect an error since we're not connected to a real k8s cluster
	// and clientset is nil
	if err == nil {
		t.Error("Read() expected error but got nil")
	}
}