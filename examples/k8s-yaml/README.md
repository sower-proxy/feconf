# K8S YAML Configuration Example

This example demonstrates how to use the conf library to read YAML configuration from a Kubernetes ConfigMap using the local ~/.kube/config file.

## Prerequisites

1. Access to a Kubernetes cluster
2. kubectl configured with ~/.kube/config
3. Permissions to create ConfigMaps in the default namespace

## Setup

1. Create the ConfigMap in your Kubernetes cluster:

```bash
kubectl create configmap my-app-config --from-file=config.yaml=./config.yaml -n default
```

2. Run the example:

```bash
go run main.go
```

## Configuration Structure

The example expects a YAML configuration with the following structure:

- app: Application settings (name, version, environment, debug)
- database: Database connection settings
- features: Feature flags and settings

## Real-time Updates

The example demonstrates real-time configuration updates by subscribing to changes in the ConfigMap. When you update the ConfigMap, the application will automatically receive the new configuration.

## URI Format

The k8s reader uses the following URI format:
```
k8s://[resource-type]/[namespace]/[name]/[key]
```

Where:
- resource-type: configmap or secret
- namespace: Kubernetes namespace
- name: Resource name
- key: Key within the resource containing the configuration