#!/bin/bash
set -e

# Function to wait for database to be ready
wait_for_db() {
    echo "Waiting for database to be ready..."
    while ! airflow db check 2>/dev/null; do
        echo "Database not ready, waiting 5 seconds..."
        sleep 5
    done
    echo "Database is ready!"
}

# Function to update Airflow configuration for newer versions
update_airflow_config() {
    local config_file="/opt/airflow/airflow.cfg"
    
    if [ -f "$config_file" ]; then
        echo "Updating Airflow configuration for newer version compatibility..."
        
        # Create backup
        cp "$config_file" "${config_file}.backup"
        
        # Update deprecated sql_alchemy_conn setting
        if grep -q "^\[core\]" "$config_file" && grep -A 50 "^\[core\]" "$config_file" | grep -q "^sql_alchemy_conn"; then
            echo "Moving sql_alchemy_conn from [core] to [database] section..."
            
            # Extract the connection string
            SQL_CONN=$(grep -A 50 "^\[core\]" "$config_file" | grep "^sql_alchemy_conn" | cut -d'=' -f2- | xargs)
            
            # Remove from [core] section
            sed -i '/^\[core\]/,/^\[/{/^sql_alchemy_conn/d;}' "$config_file"
            
            # Add to [database] section or create it
            if grep -q "^\[database\]" "$config_file"; then
                # Add after [database] section header
                sed -i "/^\[database\]/a sql_alchemy_conn = $SQL_CONN" "$config_file"
            else
                # Create [database] section
                echo -e "\n[database]\nsql_alchemy_conn = $SQL_CONN" >> "$config_file"
            fi
            
            echo "Configuration updated successfully!"
        fi
    fi
}

# Set environment variables for better error handling
export AIRFLOW__CORE__LOAD_EXAMPLES=False
export AIRFLOW__WEBSERVER__EXPOSE_CONFIG=True
export AIRFLOW__CORE__DAGS_ARE_PAUSED_AT_CREATION=True

# Only try chown if running as root user (UID 0)
if [ "$(id -u)" = "0" ]; then
    echo "Running as root, setting up permissions..."
    chown -R 50000:50000 /opt/airflow/logs /opt/airflow/dags /opt/airflow/plugins || true
    
    # Create necessary directories
    mkdir -p /opt/airflow/logs /opt/airflow/dags /opt/airflow/plugins
    chown -R 50000:50000 /opt/airflow/logs /opt/airflow/dags /opt/airflow/plugins || true
fi

# Update configuration if needed
update_airflow_config

# Handle different Airflow commands
case "$1" in
    webserver|scheduler|worker|flower)
        echo "Starting Airflow $1..."
        
        # Wait for database before starting services
        wait_for_db
        
        # Initialize database if needed (only for webserver/scheduler)
        if [ "$1" = "webserver" ] || [ "$1" = "scheduler" ]; then
            echo "Checking if database needs initialization..."
            if ! airflow db check 2>/dev/null; then
                echo "Initializing Airflow database..."
                airflow db init
            else
                # Upgrade database schema if needed
                echo "Upgrading Airflow database schema..."
                airflow db upgrade
            fi
            
            # Create default admin user if it doesn't exist
            if [ "$1" = "webserver" ]; then
                echo "Creating default admin user..."
                airflow users create \
                    --username admin \
                    --firstname Admin \
                    --lastname User \
                    --role Admin \
                    --email admin@example.com \
                    --password admin 2>/dev/null || echo "Admin user already exists"
            fi
        fi
        ;;
    db)
        echo "Running database command: $*"
        ;;
    *)
        echo "Running Airflow command: $*"
        ;;
esac

# Execute the Airflow command
exec airflow "$@"
