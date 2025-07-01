#!/bin/bash

# Airflow + Kubernetes + GO Application Setup Script
# This script sets up everything from zero to a working environment

set -e  # Exit on any error

echo "🚀 Starting Airflow + Kubernetes + GO Application Setup..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check minikube
    if ! command -v minikube &> /dev/null; then
        print_error "minikube is not installed. Please install minikube first."
        exit 1
    fi
    
    print_success "All prerequisites are installed!"
}

# Start Minikube
start_minikube() {
    print_status "Starting Minikube cluster..."
    
    if minikube status | grep -q "Running"; then
        print_warning "Minikube is already running"
    else
        minikube start --driver=docker
        print_success "Minikube started successfully!"
    fi
}

# Build GO application
build_go_app() {
    print_status "Building GO application..."
    
    cd go-sample-app
    
    # Build the Docker image
    docker build -t go-sample-app:latest .
    
    # Load image into Minikube
    minikube image load go-sample-app:latest
    
    cd ..
    
    print_success "GO application built and loaded into Minikube!"
}

# Build Airflow image
build_airflow() {
    print_status "Building Airflow image..."
    
    docker-compose build
    
    print_success "Airflow image built successfully!"
}

# Setup Kubernetes configuration
setup_kubeconfig() {
    print_status "Setting up Kubernetes configuration..."
    
    # Get Minikube kubeconfig and update it for Docker access
    minikube kubectl -- config view --minify --flatten | sed 's/127.0.0.1/host.docker.internal/g' > kubeconfig.yaml
    
    print_success "Kubernetes configuration updated!"
}

# Ensure kubeconfig is up-to-date after Minikube restarts
refresh_kubeconfig_and_restart_airflow() {
    print_status "Refreshing kubeconfig and restarting Airflow services..."
    setup_kubeconfig
    print_status "Stopping Airflow services (if running)..."
    docker-compose down 2>/dev/null || true
    print_status "Starting Airflow services..."
    docker-compose up -d
    print_success "Airflow services restarted with updated kubeconfig!"
}

# Start Airflow services
start_airflow() {
    print_status "Starting Airflow services..."
    
    # Stop any existing services
    docker-compose down 2>/dev/null || true
    
    # Start services
    docker-compose up -d
    
    print_success "Airflow services started!"
}

# Wait for Airflow to be ready
wait_for_airflow() {
    print_status "Waiting for Airflow to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/api/v1/version > /dev/null 2>&1; then
            print_success "Airflow is ready!"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Airflow not ready yet, waiting..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    print_error "Airflow failed to start within the expected time"
    return 1
}

# Clean up old pods
cleanup_pods() {
    print_status "Cleaning up old Kubernetes pods..."
    
    kubectl delete pods -n default -l kubernetes_pod_operator=True 2>/dev/null || true
    
    print_success "Old pods cleaned up!"
}

# Display final information
show_final_info() {
    echo ""
    echo "🎉 Setup Complete!"
    echo "=================="
    echo ""
    echo "📊 Airflow UI: http://localhost:8080"
    echo "   Username: airflow"
    echo "   Password: airflow"
    echo ""
    echo "🔧 Kubernetes Cluster: Minikube"
    echo "   Status: $(minikube status | grep 'host:')"
    echo ""
    echo "📁 Project Structure:"
    echo "   - dags/: Airflow DAGs"
    echo "   - go-sample-app/: GO application"
    echo "   - config/: Airflow configuration"
    echo ""
    echo "🚀 Next Steps:"
    echo "   1. Open Airflow UI at http://localhost:8080"
    echo "   2. Trigger the 'go_app_k8s_demo' DAG"
    echo "   3. Monitor the Kubernetes pods in Minikube"
    echo ""
    echo "📋 Useful Commands:"
    echo "   - View pods: kubectl get pods"
    echo "   - View logs: kubectl logs <pod-name>"
    echo "   - Stop services: docker-compose down"
    echo "   - Stop Minikube: minikube stop"
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "  Airflow + Kubernetes + GO Setup Script"
    echo "=========================================="
    echo ""
    
    check_prerequisites
    start_minikube
    refresh_kubeconfig_and_restart_airflow
    build_go_app
    build_airflow
    wait_for_airflow
    cleanup_pods
    show_final_info
}

# Run main function
main "$@" 