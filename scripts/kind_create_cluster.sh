#!/usr/bin/env sh

MACHINE_NAME="podman-machine-default"
MACHINE_STATUS="Currently running"

# Verify there's a Podman machine running
if ! podman machine list 2>/dev/null | grep -q "${MACHINE_NAME}.*${MACHINE_STATUS}"; then
    echo "Error: Podman machine '${MACHINE_NAME}' is not running."
    echo "Please start the Podman machine with: podman machine start ${MACHINE_NAME}"
    exit 1
fi

CLUSTER_NAME="kind"

# Check if a cluster with the name "kind" already exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "A Kind cluster named '${CLUSTER_NAME}' is already running. Skipping creation."
    exit 0
fi

# Create the cluster if it doesn't exist
echo "Creating Kind cluster '${CLUSTER_NAME}'..."
kind create cluster --config kind-config.yaml
