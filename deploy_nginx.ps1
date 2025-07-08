# Deploy Go Control Plane with Nginx Reverse Proxy (Windows)
# This script provides instructions for setting up nginx on Windows

param(
    [string]$Domain = "localhost"
)

Write-Host "🚀 Go Control Plane - Nginx Setup Guide (Windows)" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
Write-Host ""

Write-Host "📋 Prerequisites:" -ForegroundColor Yellow
Write-Host "1. Download nginx for Windows from http://nginx.org/en/download.html"
Write-Host "2. Extract to C:\nginx or desired location"
Write-Host "3. Ensure Go Control Plane is built and ready"
Write-Host ""

Write-Host "🔧 Configuration Steps:" -ForegroundColor Yellow
Write-Host ""

Write-Host "Step 1: Create nginx.conf in your nginx directory" -ForegroundColor Cyan
Write-Host "Copy the content from nginx.conf file in this directory" -ForegroundColor Gray
Write-Host ""

$nginxConfig = @"
# Go Storage Control Plane - Windows nginx Configuration
worker_processes 1;

events {
    worker_connections 1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;
    
    sendfile        on;
    keepalive_timeout  65;
    
    # Upstream definitions
    upstream go_auth_gateway {
        server 127.0.0.1:8090;
    }
    
    upstream go_tenant_node {
        server 127.0.0.1:8000;
    }
    
    upstream go_operation_node {
        server 127.0.0.1:8081;
    }
    
    upstream go_cbo_engine {
        server 127.0.0.1:8082;
    }
    
    upstream go_metadata_catalog {
        server 127.0.0.1:8083;
    }
    
    upstream go_monitoring {
        server 127.0.0.1:8084;
    }
    
    upstream go_query_interpreter {
        server 127.0.0.1:8085;
    }

    server {
        listen       80;
        server_name  $Domain;

        # Common proxy settings
        proxy_set_header Host `$host;
        proxy_set_header X-Real-IP `$remote_addr;
        proxy_set_header X-Forwarded-For `$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto `$scheme;

        # 🔐 Auth Gateway (Port 8090) - CORRECTED PORT
        location /auth/ {
            proxy_pass http://go_auth_gateway/;
            rewrite ^/auth/(.*)$ /`$1 break;
        }

        # 🏢 Tenant Node (Port 8000)
        location /data/ {
            proxy_pass http://go_tenant_node/;
            rewrite ^/data/(.*)$ /`$1 break;
        }

        # 🎯 Operation Node (Port 8081)
        location /query/ {
            proxy_pass http://go_operation_node/;
            rewrite ^/query/(.*)$ /`$1 break;
        }

        # 🧠 CBO Engine (Port 8082)
        location /optimize/ {
            proxy_pass http://go_cbo_engine/;
            rewrite ^/optimize/(.*)$ /`$1 break;
        }

        # 📊 Metadata Catalog (Port 8083)
        location /metadata/ {
            proxy_pass http://go_metadata_catalog/;
            rewrite ^/metadata/(.*)$ /`$1 break;
        }

        # 📈 Monitoring (Port 8084)
        location /monitor/ {
            proxy_pass http://go_monitoring/;
            rewrite ^/monitor/(.*)$ /`$1 break;
        }

        # 🔍 Query Interpreter (Port 8085)
        location /parse/ {
            proxy_pass http://go_query_interpreter/;
            rewrite ^/parse/(.*)$ /`$1 break;
        }

        # Health check
        location /health {
            proxy_pass http://go_auth_gateway/health;
        }

        # Version info
        location /version {
            proxy_pass http://go_auth_gateway/version;
        }

        # Default location
        location / {
            return 404 "Go Control Plane - Service not found";
            add_header Content-Type text/plain;
        }

        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }
    }
}
"@

# Save nginx config to file
$nginxConfig | Out-File -FilePath "nginx-windows.conf" -Encoding UTF8

Write-Host "✅ Created nginx-windows.conf" -ForegroundColor Green
Write-Host ""

Write-Host "Step 2: Start Services" -ForegroundColor Cyan
Write-Host "In PowerShell (Admin):" -ForegroundColor Gray
Write-Host ""
Write-Host "# Start Go Control Plane" -ForegroundColor Gray
Write-Host "go run ." -ForegroundColor White
Write-Host ""
Write-Host "# In another terminal, start nginx" -ForegroundColor Gray
Write-Host "cd C:\nginx" -ForegroundColor White
Write-Host ".\nginx.exe" -ForegroundColor White
Write-Host ""

Write-Host "Step 3: Test Configuration" -ForegroundColor Cyan

# Create test script
$testScript = @"
# Test Go Control Plane via Nginx
# Run this after starting both services

Write-Host "🧪 Testing Go Control Plane via Nginx" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

`$baseUrl = "http://$Domain"
`$endpoints = @(
    "/health",
    "/auth/login",
    "/data/stats", 
    "/query/status",
    "/optimize/stats",
    "/metadata/stats",
    "/monitor/metrics",
    "/parse/health"
)

foreach (`$endpoint in `$endpoints) {
    `$url = "`$baseUrl`$endpoint"
    try {
        `$response = Invoke-WebRequest -Uri `$url -Method GET -TimeoutSec 5
        if (`$response.StatusCode -eq 200) {
            Write-Host "✅ `$endpoint - OK" -ForegroundColor Green
        } else {
            Write-Host "⚠️  `$endpoint - `$(`$response.StatusCode)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "❌ `$endpoint - Failed: `$(`$_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🌐 Access URLs:" -ForegroundColor Yellow
Write-Host "  Main Health: http://$Domain/health"
Write-Host "  Auth API:    http://$Domain/auth/"
Write-Host "  Data API:    http://$Domain/data/"
Write-Host "  Query API:   http://$Domain/query/"
Write-Host "  Monitor:     http://$Domain/monitor/"
"@

$testScript | Out-File -FilePath "test_nginx.ps1" -Encoding UTF8

Write-Host "✅ Created test_nginx.ps1" -ForegroundColor Green
Write-Host ""

Write-Host "Step 4: Management Commands" -ForegroundColor Cyan
Write-Host ""
Write-Host "Stop nginx:" -ForegroundColor Gray
Write-Host "cd C:\nginx && .\nginx.exe -s stop" -ForegroundColor White
Write-Host ""
Write-Host "Reload nginx config:" -ForegroundColor Gray  
Write-Host "cd C:\nginx && .\nginx.exe -s reload" -ForegroundColor White
Write-Host ""
Write-Host "Test nginx config:" -ForegroundColor Gray
Write-Host "cd C:\nginx && .\nginx.exe -t" -ForegroundColor White
Write-Host ""

Write-Host "🎉 Setup Complete!" -ForegroundColor Green
Write-Host "=================" -ForegroundColor Green
Write-Host ""
Write-Host "Files created:" -ForegroundColor Yellow
Write-Host "  ✅ nginx-windows.conf - Nginx configuration"
Write-Host "  ✅ test_nginx.ps1 - Test script"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Copy nginx-windows.conf to your nginx directory as nginx.conf"
Write-Host "2. Start Go Control Plane: go run ."
Write-Host "3. Start nginx: .\nginx.exe"
Write-Host "4. Test: .\test_nginx.ps1"
Write-Host ""
Write-Host "Access your services at: http://$Domain/" -ForegroundColor Green
