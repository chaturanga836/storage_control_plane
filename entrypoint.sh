#!/bin/bash
# entrypoint.sh

# Fix permissions automatically
chown -R 50000:50000 /opt/airflow/logs /opt/airflow/dags /opt/airflow/plugins

# Run airflow command passed as args
exec "$@"
