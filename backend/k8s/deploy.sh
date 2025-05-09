#!/bin/bash
# deploy.sh - Deploy the Alya.io backend to Kubernetes

# Set default environment
ENV=${1:-development}
REGISTRY=${REGISTRY:-localhost:5000}
TAG=${TAG:-latest}

# Display info
echo "Deploying Alya.io backend to Kubernetes"
echo "Environment: $ENV"
echo "Registry: $REGISTRY"
echo "Tag: $TAG"

# Ensure kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found"
    exit 1
fi

# Ensure kustomize is available
if ! command -v kustomize &> /dev/null; then
    echo "Kustomize not found, using kubectl kustomize"
    KUSTOMIZE="kubectl kustomize"
else
    KUSTOMIZE="kustomize"
fi

# Deploy dependencies first
echo "Deploying dependencies..."
$KUSTOMIZE k8s/dependencies | kubectl apply -f -

# Wait for dependencies to be available
echo "Waiting for dependencies to be ready..."
kubectl wait --for=condition=ready pod -l app=alya-postgres --timeout=120s -n alya
kubectl wait --for=condition=ready pod -l app=alya-redis --timeout=60s -n alya

# Deploy base with environment overlay
echo "Deploying application for $ENV environment..."
$KUSTOMIZE k8s/overlays/$ENV | kubectl apply -f -

# Deploy ingress if needed
if [ "$ENV" = "production" ]; then
    echo "Deploying ingress..."
    kubectl apply -f k8s/ingress.yaml
fi

echo "Deployment completed!"
echo "Checking deployment status..."
kubectl get pods -n alya