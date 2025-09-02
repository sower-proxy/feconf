#!/bin/bash

# Example script to create and update ConfigMap with variations for demonstration

set -e

CONFIGMAP_NAME="my-app-config"
NAMESPACE="default"
CONFIG_FILE="config.yaml"

echo "=== Kubernetes ConfigMap Configuration Example ==="
echo

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Generate timestamp for tracking
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Function to generate config variations
generate_config() {
    local version=$1
    local debug=$2
    local max_conn=$3
    local enable_cache=$4
    
    cat > "$CONFIG_FILE" << EOF
app:
  name: "my-awesome-app"
  version: "$version"
  environment: "development"
  debug: $debug

database:
  host: "postgres-service"
  port: 5432
  name: "myapp_db"
  max_connections: $max_conn
  connection_timeout: 30

features:
  enable_cache: $enable_cache
  enable_metrics: true
  allowed_origins:
    - "https://example.com"
    - "https://api.example.com"
    - "http://localhost:3000"
    - "https://staging.example.com"  # Added at $TIMESTAMP
EOF
}

# Check if ConfigMap already exists
if kubectl get configmap "$CONFIGMAP_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo "ConfigMap '$CONFIGMAP_NAME' already exists in namespace '$NAMESPACE'"
    
    # Get current update count or initialize to 0
    UPDATE_COUNT=$(kubectl get configmap "$CONFIGMAP_NAME" -n "$NAMESPACE" -o jsonpath='{.metadata.annotations.update-count}' 2>/dev/null || echo "0")
    UPDATE_COUNT=$((UPDATE_COUNT + 1))
    
    echo "This is update #$UPDATE_COUNT"
    echo
    
    # Generate variations based on update count
    case $((UPDATE_COUNT % 4)) in
        0)
            echo "=== Update $UPDATE_COUNT: Performance Mode ==="
            generate_config "1.$UPDATE_COUNT.0" false 50 true
            ;;
        1)
            echo "=== Update $UPDATE_COUNT: Debug Mode ==="
            generate_config "1.$UPDATE_COUNT.0" true 10 false
            ;;
        2)
            echo "=== Update $UPDATE_COUNT: Balanced Mode ==="
            generate_config "1.$UPDATE_COUNT.0" true 25 true
            ;;
        3)
            echo "=== Update $UPDATE_COUNT: Production Mode ==="
            generate_config "1.$UPDATE_COUNT.0" false 30 false
            ;;
    esac
    
    echo "Configuration updated with:"
    echo "  Version: 1.$UPDATE_COUNT.0"
    echo "  Debug: $([ $((UPDATE_COUNT % 4)) -ne 0 ] && [ $((UPDATE_COUNT % 4)) -ne 3 ] && echo "true" || echo "false")"
    echo "  Max Connections: $([ $((UPDATE_COUNT % 4)) -eq 0 ] && echo "50" || ([ $((UPDATE_COUNT % 4)) -eq 1 ] && echo "10" || ([ $((UPDATE_COUNT % 4)) -eq 2 ] && echo "25" || echo "30")))"
    echo "  Cache Enabled: $([ $((UPDATE_COUNT % 4)) -eq 0 ] || [ $((UPDATE_COUNT % 4)) -eq 2 ] && echo "true" || echo "false")"
    echo
    
    # Update ConfigMap
    kubectl create configmap "$CONFIGMAP_NAME" \
        --from-file="config.yaml=$CONFIG_FILE" \
        -n "$NAMESPACE" \
        --dry-run=client -o yaml | \
    kubectl apply -f -
    
    # Add annotations separately
    kubectl annotate configmap "$CONFIGMAP_NAME" \
        -n "$NAMESPACE" \
        --overwrite \
        update-count="$UPDATE_COUNT" \
        last-update="$TIMESTAMP"
    
else
    echo "Creating new ConfigMap '$CONFIGMAP_NAME' in namespace '$NAMESPACE'..."
    echo "=== Initial Configuration ==="
    
    # Generate initial config
    generate_config "1.0.0" true 20 true
    
    # Create ConfigMap
    kubectl create configmap "$CONFIGMAP_NAME" \
        --from-file="config.yaml=$CONFIG_FILE" \
        -n "$NAMESPACE"
    
    # Add annotations
    kubectl annotate configmap "$CONFIGMAP_NAME" \
        -n "$NAMESPACE" \
        --overwrite \
        update-count="0" \
        last-update="$TIMESTAMP" \
        created-by="setup.sh"
fi

echo
echo "ConfigMap created/updated successfully!"
echo
echo "To view the ConfigMap:"
echo "  kubectl get configmap $CONFIGMAP_NAME -n $NAMESPACE -o yaml"
echo
echo "To run the example application:"
echo "  go run main.go"
echo
echo "To test updates, run this script again:"
echo "  ./setup.sh"
echo
echo "Each run will cycle through different configuration modes:"
echo "  1. Performance Mode (high connections, cache on)"
echo "  2. Debug Mode (low connections, cache off)"
echo "  3. Balanced Mode (medium connections, cache on)"
echo "  4. Production Mode (medium connections, cache off)"