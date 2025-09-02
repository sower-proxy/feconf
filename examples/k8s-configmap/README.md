# Kubernetes ConfigMap/Secret Reader Example

This example demonstrates how to use the Kubernetes ConfigMap/Secret reader with the conf library.

## URI Format

The Kubernetes reader supports the following URI format:

```
k8s://{resourceType}/{namespace}/{name}[/{key}]
```

Where:
- `resourceType`: Either `configmap` or `secret`
- `namespace`: The Kubernetes namespace
- `name`: The name of the ConfigMap or Secret
- `key`: Optional key within the ConfigMap or Secret

Examples:
- `k8s://configmap/default/app-config` - Reads a ConfigMap, returns first key's value
- `k8s://configmap/default/app-config/config.yaml` - Reads a specific key from a ConfigMap
- `k8s://secret/default/db-secret` - Reads a Secret, returns first key's value
- `k8s://secret/default/db-secret/password` - Reads a specific key from a Secret

## Prerequisites

1. A Kubernetes cluster
2. A ConfigMap or Secret in the cluster

Example ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
    database:
      host: "localhost"
      port: 5432
      username: "user"
      password: "pass"
```

Example Secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
  namespace: default
type: Opaque
data:
  username: "dXNlcg=="  # base64 encoded "user"
  password: "cGFzcw=="  # base64 encoded "pass"
```

## Running the Example

1. Ensure you have a Kubernetes cluster accessible
2. Create the example ConfigMap/Secret in your cluster
3. Set up your kubeconfig (usually at `~/.kube/config`)
4. Run the example:

```bash
cd examples/k8s-configmap
go run main.go
```

## Authentication

The Kubernetes reader supports two authentication methods:

1. **In-cluster**: When running inside a Kubernetes pod, it will automatically use the service account token
2. **Kubeconfig**: When running outside the cluster, it will use the kubeconfig file (default: `~/.kube/config`)

You can also specify a custom kubeconfig path using the `KUBECONFIG` environment variable:

```bash
KUBECONFIG=/path/to/kubeconfig go run main.go
```