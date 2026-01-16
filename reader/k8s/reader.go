// Package k8s provides a configuration reader implementation for Kubernetes ConfigMaps and Secrets.
// It supports reading configuration data from Kubernetes resources and subscribing to changes.
package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sower-proxy/feconf/reader"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// SchemeK8S represents kubernetes URI scheme
	SchemeK8S reader.Scheme = "k8s"

	// ResourceTypeConfigMap represents configmap resource type
	ResourceTypeConfigMap = "configmap"

	// ResourceTypeSecret represents secret resource type
	ResourceTypeSecret = "secret"
)

// init registers k8s reader
func init() {
	_ = reader.RegisterReader(SchemeK8S, func(uri string) (reader.ConfReader, error) {
		return NewK8SReader(uri)
	})
}

// K8SReader implements ConfReader for kubernetes configmap/secret
type K8SReader struct {
	uri          string
	resourceType string
	namespace    string
	name         string
	key          string
	clientset    kubernetes.Interface
	informer     cache.SharedIndexInformer
	stopCh       chan struct{}
	mu           sync.RWMutex
	closed       bool
}

// NewK8SReader creates a new k8s reader
// URI format: k8s://{resourceType}/{namespace}/{name}[/{key}]
// Example: k8s://configmap/default/my-config/config.yaml
func NewK8SReader(uri string) (*K8SReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %w", err)
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
	// Using strings.Split to properly handle path segments
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid k8s URI path, expected format: k8s://{resourceType}/{namespace}/{name}[/{key}]")
	}

	namespace := parts[0]
	name := parts[1]
	key := ""
	if len(parts) > 2 {
		key = parts[2]
	}

	// Create k8s client
	clientset, err := createK8SClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &K8SReader{
		uri:          uri,
		resourceType: resourceType,
		namespace:    namespace,
		name:         name,
		key:          key,
		clientset:    clientset,
	}, nil
}

// Read reads configuration data from k8s configmap/secret
func (k *K8SReader) Read(ctx context.Context) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check if clientset is nil (for testing)
	if k.clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	switch k.resourceType {
	case ResourceTypeConfigMap:
		return k.readConfigMap(ctx)
	case ResourceTypeSecret:
		return k.readSecret(ctx)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", k.resourceType)
	}
}

// Subscribe subscribes to k8s configmap/secret changes and returns update channel
func (k *K8SReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	if k.informer != nil {
		return nil, fmt.Errorf("already subscribed")
	}

	// Check if clientset is nil (for testing)
	if k.clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	// Create informer based on resource type
	factory := informers.NewSharedInformerFactoryWithOptions(
		k.clientset, 0, informers.WithNamespace(k.namespace))

	var informer cache.SharedIndexInformer
	switch k.resourceType {
	case ResourceTypeConfigMap:
		informer = factory.Core().V1().ConfigMaps().Informer()
	case ResourceTypeSecret:
		informer = factory.Core().V1().Secrets().Informer()
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", k.resourceType)
	}

	k.informer = informer
	k.stopCh = make(chan struct{})

	eventChan := make(chan *reader.ReadEvent, 1)

	// Set up event handlers
	_, _ = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			k.handleResourceUpdate(ctx, eventChan)
		},
		UpdateFunc: func(oldObj, newObj any) {
			k.handleResourceUpdate(ctx, eventChan)
		},
		DeleteFunc: func(obj any) {
			// Send empty event when resource is deleted
			confEvent := reader.NewReadEvent(k.uri, nil, fmt.Errorf("%s deleted", k.resourceType))
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				// Context cancelled, ignore the event
			}
		},
	})

	// Start informer
	go informer.Run(k.stopCh)

	// Wait for cache sync
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		// Clean up resources on failure
		close(k.stopCh)
		k.informer = nil
		k.stopCh = nil
		return nil, fmt.Errorf("failed to sync cache")
	}

	return eventChan, nil
}

// Close closes the reader and cleans up resources
func (k *K8SReader) Close() error {
	if k == nil {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return nil
	}

	k.closed = true

	if k.stopCh != nil {
		close(k.stopCh)
		k.stopCh = nil
	}

	k.informer = nil

	return nil
}

// readConfigMap reads data from configmap
func (k *K8SReader) readConfigMap(ctx context.Context) ([]byte, error) {
	// Check if clientset is nil (for testing)
	if k.clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	cm, err := k.clientset.CoreV1().ConfigMaps(k.namespace).Get(ctx, k.name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", k.namespace, k.name, err)
	}

	if k.key != "" {
		// Return specific key value
		if value, exists := cm.Data[k.key]; exists {
			return []byte(value), nil
		}
		return nil, fmt.Errorf("key %s not found in configmap %s/%s", k.key, k.namespace, k.name)
	}

	// When no specific key is requested, return the first key's value if there's exactly one
	if len(cm.Data) == 0 {
		return nil, fmt.Errorf("configmap %s/%s is empty", k.namespace, k.name)
	}

	if len(cm.Data) == 1 {
		// Return the value of the single key
		for _, value := range cm.Data {
			return []byte(value), nil
		}
	}

	// Multiple keys exist but no specific key requested
	return nil, fmt.Errorf("configmap %s/%s contains multiple keys, please specify one: %v", k.namespace, k.name, getMapKeys(cm.Data))
}

// readSecret reads data from secret
func (k *K8SReader) readSecret(ctx context.Context) ([]byte, error) {
	// Check if clientset is nil (for testing)
	if k.clientset == nil {
		return nil, fmt.Errorf("k8s client not initialized")
	}

	secret, err := k.clientset.CoreV1().Secrets(k.namespace).Get(ctx, k.name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", k.namespace, k.name, err)
	}

	if k.key != "" {
		// Return specific key value
		if value, exists := secret.Data[k.key]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("key %s not found in secret %s/%s", k.key, k.namespace, k.name)
	}

	// When no specific key is requested, return the first key's value if there's exactly one
	if len(secret.Data) == 0 {
		return nil, fmt.Errorf("secret %s/%s is empty", k.namespace, k.name)
	}

	if len(secret.Data) == 1 {
		// Return the value of the single key
		for _, value := range secret.Data {
			return value, nil
		}
	}

	// Multiple keys exist but no specific key requested
	return nil, fmt.Errorf("secret %s/%s contains multiple keys, please specify one: %v", k.namespace, k.name, getMapKeysByte(secret.Data))
}

// handleResourceUpdate handles resource update events
func (k *K8SReader) handleResourceUpdate(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	// Add small delay to ensure resource update is complete
	time.Sleep(100 * time.Millisecond)

	data, err := k.Read(ctx)
	confEvent := reader.NewReadEvent(k.uri, data, err)

	select {
	case eventChan <- confEvent:
	case <-ctx.Done():
		// Context cancelled, ignore the event
	}
}

// createK8SClient creates k8s client
func createK8SClient() (kubernetes.Interface, error) {
	// Check if we're running in-cluster
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		// Use in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create k8s client from in-cluster config: %w", err)
		}
		return clientset, nil
	}

	// Use kubeconfig from environment or default location
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		kubeconfig = homeDir + "/.kube/config"
	}

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig (%s): %w", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client from kubeconfig: %w", err)
	}
	return clientset, nil
}

// getMapKeys returns keys from string map
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMapKeysByte returns keys from byte slice map
func getMapKeysByte(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
