from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.bash import BashOperator
import sys
import platform
import subprocess
import os

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2024, 1, 1),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

def print_worker_info():
    """Print worker image and system information"""
    print("=" * 50)
    print("WORKER IMAGE INFORMATION")
    print("=" * 50)
    
    # Python information
    print(f"Python Version: {sys.version}")
    print(f"Python Executable: {sys.executable}")
    
    # System information
    print(f"Platform: {platform.platform()}")
    print(f"System: {platform.system()}")
    print(f"Machine: {platform.machine()}")
    print(f"Processor: {platform.processor()}")
    
    # Environment information
    print(f"Hostname: {platform.node()}")
    print(f"Current Working Directory: {os.getcwd()}")
    print(f"User: {os.getenv('USER', 'Unknown')}")
    print(f"UID: {os.getuid() if hasattr(os, 'getuid') else 'N/A'}")
    
    # Docker/Container information
    print(f"Container ID: {os.getenv('HOSTNAME', 'Unknown')}")
    print(f"AIRFLOW_HOME: {os.getenv('AIRFLOW_HOME', 'Not set')}")
    print(f"AIRFLOW_UID: {os.getenv('AIRFLOW_UID', 'Not set')}")
    
    # Check if running in Docker
    if os.path.exists('/.dockerenv'):
        print("Running in Docker container: Yes")
    else:
        print("Running in Docker container: No")
    
    print("=" * 50)

def print_installed_packages():
    """Print list of installed Python packages"""
    print("=" * 50)
    print("INSTALLED PACKAGES")
    print("=" * 50)
    
    try:
        result = subprocess.run([sys.executable, '-m', 'pip', 'list'], 
                              capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            print(result.stdout)
        else:
            print(f"Error getting packages: {result.stderr}")
    except Exception as e:
        print(f"Exception getting packages: {e}")
    
    print("=" * 50)

def print_airflow_info():
    """Print Airflow specific information"""
    print("=" * 50)
    print("AIRFLOW INFORMATION")
    print("=" * 50)
    
    try:
        import airflow
        print(f"Airflow Version: {airflow.__version__}")
    except ImportError:
        print("Airflow not available")
    
    # Check for asyncpg specifically
    try:
        import asyncpg
        print(f"asyncpg Version: {asyncpg.__version__}")
        print("✅ asyncpg is available!")
    except ImportError:
        print("❌ asyncpg is NOT available")
    
    print("=" * 50)

# Create the DAG
dag = DAG(
    'sample_worker_info',
    default_args=default_args,
    description='Sample DAG to print worker image information',
    schedule=timedelta(days=1),
    catchup=False,
    tags=['sample', 'debug', 'worker-info'],
)

# Define tasks
task_print_worker_info = PythonOperator(
    task_id='print_worker_info',
    python_callable=print_worker_info,
    dag=dag,
)

task_print_packages = PythonOperator(
    task_id='print_installed_packages',
    python_callable=print_installed_packages,
    dag=dag,
)

task_print_airflow_info = PythonOperator(
    task_id='print_airflow_info',
    python_callable=print_airflow_info,
    dag=dag,
)

task_system_info = BashOperator(
    task_id='system_info',
    bash_command="""
    echo "=== SYSTEM INFORMATION ==="
    echo "Date: $(date)"
    echo "Uptime: $(uptime)"
    echo "Memory: $(free -h)"
    echo "Disk: $(df -h /)"
    echo "CPU Info: $(cat /proc/cpuinfo | grep 'model name' | head -1)"
    echo "=== DOCKER INFO ==="
    if command -v docker &> /dev/null; then
        echo "Docker version: $(docker --version)"
    else
        echo "Docker not available"
    fi
    echo "=== ENVIRONMENT VARIABLES ==="
    env | grep -E "(AIRFLOW|PYTHON|PATH)" | sort
    """,
    dag=dag,
)

# Set task dependencies
task_print_worker_info >> task_print_packages >> task_print_airflow_info >> task_system_info
