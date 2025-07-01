"""
DAG to run GO application using KubernetesPodOperator
"""
from datetime import datetime, timedelta
from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator
from airflow.providers.cncf.kubernetes.secret import Secret
from kubernetes.client import models as k8s_models

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2024, 1, 1),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': None,
    'retry_delay': timedelta(minutes=5),
}

# Create the DAG
dag = DAG(
    'go_app_k8s_demo',
    default_args=default_args,
    description='Run GO application using KubernetesPodOperator',
    schedule=None,
    catchup=False,
    tags=['go', 'kubernetes', 'demo'],
)

# Task 1: Run GO app with local config
run_go_app_local_config = KubernetesPodOperator(
    task_id='run_go_app_local_config',
    name='go-app-local-config',
    namespace='default',
    image='go-sample-app:latest',
    image_pull_policy='Never',
    cmds=['./main'],
    arguments=['--config-type=local'],
    env_vars=[
        k8s_models.V1EnvVar(name='CONFIG_TYPE', value='local'),
        k8s_models.V1EnvVar(name='CONFIG_PATH', value='/app/config/config.yaml'),
    ],
    get_logs=True,
    is_delete_operator_pod=True,
    dag=dag,
)

# Task 2: Run GO app with environment variables
run_go_app_env_config = KubernetesPodOperator(
    task_id='run_go_app_env_config',
    name='go-app-env-config',
    namespace='default',
    image='go-sample-app:latest',
    image_pull_policy='Never',
    cmds=['./main'],
    arguments=['--config-type=env'],
    env_vars=[
        k8s_models.V1EnvVar(name='CONFIG_TYPE', value='env'),
        k8s_models.V1EnvVar(name='APP_NAME', value='Go Sample App'),
        k8s_models.V1EnvVar(name='APP_VERSION', value='1.0.0'),
        k8s_models.V1EnvVar(name='APP_PORT', value='8080'),
        k8s_models.V1EnvVar(name='APP_ENVIRONMENT', value='development'),
    ],
    get_logs=True,
    is_delete_operator_pod=True,
    dag=dag,
)

# Task 3: Run GO app with remote config
run_go_app_remote_config = KubernetesPodOperator(
    task_id='run_go_app_remote_config',
    name='go-app-remote-config',
    namespace='default',
    image='go-sample-app:latest',
    image_pull_policy='Never',
    cmds=['./main'],
    arguments=['--config-type=remote'],
    env_vars=[
        k8s_models.V1EnvVar(name='CONFIG_TYPE', value='remote'),
        k8s_models.V1EnvVar(name='CONFIG_REMOTE_URL', value='https://raw.githubusercontent.com/nvduc91/af_multi_lang_poc/main/go-sample-app/config/config.yaml'),
    ],
    get_logs=True,
    is_delete_operator_pod=True,
    dag=dag,
)

# Task 4: Run GO app with custom resource limits
run_go_app_with_resources = KubernetesPodOperator(
    task_id='run_go_app_with_resources',
    name='go-app-with-resources',
    namespace='default',
    image='go-sample-app:latest',
    image_pull_policy='Never',
    cmds=['./main'],
    arguments=['--config-type=env'],
    env_vars=[
        k8s_models.V1EnvVar(name='CONFIG_TYPE', value='env'),
        k8s_models.V1EnvVar(name='APP_NAME', value='Go App with Resources'),
    ],
    container_resources=k8s_models.V1ResourceRequirements(
        limits={"memory": "512Mi", "cpu": "500m"},
        requests={"memory": "256Mi", "cpu": "250m"}
    ),
    get_logs=True,
    is_delete_operator_pod=True,
    dag=dag,
)

# Define task dependencies
run_go_app_local_config >> run_go_app_env_config >> run_go_app_remote_config >> run_go_app_with_resources 