package k8s

import (
	"context"
	"testing"

	"github.com/sower-proxy/feconf/reader"
)

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
			target, err := parseK8STarget(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseK8STarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if target.resourceType != tt.resourceType {
					t.Errorf("parseK8STarget() resourceType = %v, want %v", target.resourceType, tt.resourceType)
				}
				if target.namespace != tt.namespace {
					t.Errorf("parseK8STarget() namespace = %v, want %v", target.namespace, tt.namespace)
				}
				if target.name != tt.configName {
					t.Errorf("parseK8STarget() name = %v, want %v", target.name, tt.configName)
				}
				if target.key != tt.key {
					t.Errorf("parseK8STarget() key = %v, want %v", target.key, tt.key)
				}
			}
		})
	}
}

func TestK8SReaderInterface(t *testing.T) {
	var _ reader.ConfReader = &K8SReader{}
}

func TestK8SReaderReadWithoutClient(t *testing.T) {
	r := &K8SReader{
		resourceType: ResourceTypeConfigMap,
		namespace:    "default",
		name:         "test-config",
	}

	ctx := context.Background()
	_, err := r.Read(ctx)
	if err == nil {
		t.Error("Read() expected error but got nil")
	}
}
