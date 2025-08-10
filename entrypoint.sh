#!/bin/bash
set -e

# Only try chown if running as root user (UID 0)
if [ "$(id -u)" = "0" ]; then
  chown -R 50000:50000 /opt/airflow/logs /opt/airflow/dags /opt/airflow/plugins || true
fi

exec airflow "$@"
