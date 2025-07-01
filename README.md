# Airflow + Kubernetes + GO Application Integration

This project demonstrates how to integrate Apache Airflow with Kubernetes to run GO applications as scheduled tasks. The setup includes a local Minikube cluster, Airflow running in Docker Compose, and a GO application that can be executed as Kubernetes pods.

## 🏗️ Architecture

```
poc_af/
├── dags/                    # Airflow DAGs
│   └── dag_go_k8s.py       # DAG that runs GO app in Kubernetes
├── go-sample-app/          # GO application
│   ├── main.go             # Main GO application
│   ├── Dockerfile          # GO app Dockerfile
│   ├── go.mod              # GO dependencies
│   └── config/             # Configuration files
├── config/                 # Airflow configuration
│   └── airflow.cfg         # Airflow settings
├── docker-compose.yaml     # Airflow services
├── Dockerfile              # Airflow image with providers
├── kubeconfig.yaml         # Kubernetes configuration
├── run.sh                  # Complete setup script
└── README.md               # This file
```

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose**: For running Airflow services
- **Minikube**: For local Kubernetes cluster
- **kubectl**: Kubernetes command-line tool
- **Git**: For version control

### Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd poc_af
   ```

2. **Run the setup script**:
   ```bash
   ./run.sh
   ```

   This script will:
   - Check prerequisites
   - Start Minikube cluster
   - Build GO application
   - Build Airflow image
   - Configure Kubernetes access
   - Start Airflow services
   - Wait for services to be ready

3. **Access Airflow UI**:
   - URL: http://localhost:8080
   - Username: `airflow`
   - Password: `airflow`

## 📋 Manual Setup (Alternative)

If you prefer to run steps manually:

### 1. Start Minikube
```bash
minikube start --driver=docker
```

### 2. Build GO Application
```bash
cd go-sample-app
docker build -t go-sample-app:latest .
minikube image load go-sample-app:latest
cd ..
```

### 3. Setup Kubernetes Configuration
```bash
minikube kubectl -- config view --minify --flatten | sed 's/127.0.0.1/host.docker.internal/g' > kubeconfig.yaml
```

### 4. Build and Start Airflow
```bash
docker-compose build
docker-compose up -d
```

## 🔧 Configuration

### Airflow Configuration

The Airflow configuration is in `config/airflow.cfg`:
- **Executor**: CeleryExecutor
- **Kubernetes**: External cluster (Minikube)
- **SSL Verification**: Disabled for local development

### GO Application Configuration

The GO application supports multiple configuration sources:
- **Local**: Reads from `/app/config/config.yaml`
- **Environment**: Uses environment variables
- **Remote**: Downloads from GitHub URLs

### Kubernetes Configuration

- **Cluster**: Minikube
- **Namespace**: `default`
- **Image Pull Policy**: `Never` (uses local images)

## 📊 DAGs

### `go_app_k8s_demo`

This DAG demonstrates running GO applications in Kubernetes with different configurations:

1. **Task 1**: `run_go_app_local_config`
   - Loads configuration from local file
   - Uses local config.yaml

2. **Task 2**: `run_go_app_env_config`
   - Uses environment variables for configuration
   - Sets app name, version, port, environment

3. **Task 3**: `run_go_app_remote_config`
   - Loads configuration from remote GitHub URL
   - Downloads config file at runtime

4. **Task 4**: `run_go_app_with_resources`
   - Runs with custom resource limits
   - Demonstrates resource management

## 🛠️ Development

### Adding New DAGs

1. Create a new Python file in `dags/`
2. Use `KubernetesPodOperator` to run GO applications
3. Example:
   ```python
   from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator
   
   task = KubernetesPodOperator(
       task_id='my_go_task',
       name='my-go-app',
       namespace='default',
       image='go-sample-app:latest',
       image_pull_policy='Never',
       cmds=['./main'],
       arguments=['--config-type=env'],
       env_vars=[...],
       get_logs=True,
       is_delete_operator_pod=True,
       dag=dag,
   )
   ```

### Modifying GO Application

1. Edit files in `go-sample-app/`
2. Rebuild the image:
   ```bash
   cd go-sample-app
   docker build -t go-sample-app:latest .
   minikube image load go-sample-app:latest
   cd ..
   ```

### Adding New Configuration Types

1. Modify `go-sample-app/main.go`
2. Add new case in the `loadConfig()` function
3. Rebuild the application

## 🔍 Monitoring

### View Kubernetes Pods
```bash
kubectl get pods
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### View Airflow Logs
```bash
docker-compose logs airflow-scheduler
docker-compose logs airflow-worker
```

### Access Minikube Dashboard
```bash
minikube dashboard
```

## 🧹 Cleanup

### Stop Services
```bash
docker-compose down
```

### Stop Minikube
```bash
minikube stop
```

### Complete Cleanup
```bash
docker-compose down -v  # Remove volumes
minikube delete         # Delete cluster
```

## 🐛 Troubleshooting

### Common Issues

1. **Kubernetes Connection Failed**
   - Ensure Minikube is running: `minikube status`
   - Check kubeconfig: `cat kubeconfig.yaml`

2. **Image Pull Errors**
   - Rebuild and reload image:
     ```bash
     cd go-sample-app
     docker build -t go-sample-app:latest .
     minikube image load go-sample-app:latest
     cd ..
     ```

3. **Airflow Not Starting**
   - Check logs: `docker-compose logs airflow-scheduler`
   - Ensure ports are available: `lsof -i :8080`

4. **SSL Certificate Errors**
   - This is expected in local development
   - SSL verification is disabled in `airflow.cfg`

### Debug Commands

```bash
# Check Minikube status
minikube status

# Check Kubernetes nodes
kubectl get nodes

# Check Airflow services
docker-compose ps

# Check Airflow DAGs
curl http://localhost:8080/api/v1/dags

# Test Kubernetes connection
kubectl get pods
```

## 📚 Additional Resources

- [Apache Airflow Documentation](https://airflow.apache.org/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [GO Documentation](https://golang.org/doc/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with `./run.sh`
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
