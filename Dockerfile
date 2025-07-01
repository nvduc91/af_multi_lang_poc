FROM apache/airflow:3.0.0

# Install asyncpg and other dependencies as airflow user
USER airflow
RUN pip install --no-cache-dir \
    asyncpg \
    apache-airflow-providers-fab \
    apache-airflow-providers-celery \
    apache-airflow-providers-postgres \
    apache-airflow-providers-docker \
    apache-airflow-providers-kubernetes 