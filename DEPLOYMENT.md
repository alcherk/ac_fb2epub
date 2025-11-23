# Deployment Guide - FB2 to EPUB Converter Service

This guide covers deploying the FB2 to EPUB converter service on a fresh VPS server running Ubuntu/Debian.

## Deployment Options

1. **Docker Deployment** (Recommended) - Easy, containerized deployment
2. **Native Deployment** - Traditional binary deployment

## Prerequisites

- Fresh Ubuntu 20.04+ or Debian 11+ VPS
- Root or sudo access
- SSH access to the server
- Domain name (optional, for production)

## Option 1: Docker Deployment (Recommended)

### 1.1 Install Docker and Docker Compose

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### 1.2 Clone Repository

```bash
cd /opt
sudo git clone <your-repo-url> fb2epub
cd fb2epub
```

### 1.3 Configure Environment

Edit `docker-compose.yml` if needed, or create `.env` file:

```bash
sudo nano docker-compose.yml
```

### 1.4 Build and Run

```bash
# Build and start
sudo docker-compose up -d

# View logs
sudo docker-compose logs -f

# Check status
sudo docker-compose ps
```

### 1.5 Setup Nginx Reverse Proxy

Follow Step 6 in the Native Deployment section below, but point to `localhost:8080` (Docker container port).

### 1.6 Setup SSL

Follow Step 7 in the Native Deployment section below.

### 1.7 Maintenance

```bash
# Restart service
sudo docker-compose restart

# Update application
cd /opt/fb2epub
sudo git pull
sudo docker-compose up -d --build

# View logs
sudo docker-compose logs -f fb2epub

# Stop service
sudo docker-compose down
```

## Option 2: Native Deployment

## Step 1: Initial Server Setup

### 1.1 Update System Packages

```bash
sudo apt update
sudo apt upgrade -y
```

### 1.2 Install Required Packages

```bash
sudo apt install -y curl wget git build-essential
```

### 1.3 Install Go

```bash
# Download Go (check latest version at https://golang.org/dl/)
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# Remove old Go installation if exists
sudo rm -rf /usr/local/go

# Extract and install
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### 1.4 Create Application User

```bash
sudo useradd -m -s /bin/bash fb2epub
sudo mkdir -p /opt/fb2epub
sudo chown fb2epub:fb2epub /opt/fb2epub
```

## Step 2: Deploy Application

### 2.1 Clone Repository (or upload files)

**Option A: If using Git**
```bash
sudo -u fb2epub git clone <your-repo-url> /opt/fb2epub
cd /opt/fb2epub
```

**Option B: If uploading files**
```bash
# On your local machine, build for Linux
make linux-build
# or
GOOS=linux GOARCH=amd64 go build -o fb2epub-linux

# Upload files to server using scp
scp -r . fb2epub@your-server:/opt/fb2epub/
```

### 2.2 Build Application on Server

```bash
cd /opt/fb2epub
sudo -u fb2epub go mod download
sudo -u fb2epub go build -o fb2epub
```

### 2.3 Create Application Directories

```bash
sudo -u fb2epub mkdir -p /opt/fb2epub/{logs,temp}
sudo chmod 755 /opt/fb2epub/temp
```

## Step 3: Configure Application

### 3.1 Create Environment File

```bash
sudo -u fb2epub cat > /opt/fb2epub/.env << EOF
PORT=8080
ENVIRONMENT=production
TEMP_DIR=/opt/fb2epub/temp
MAX_FILE_SIZE=52428800
EOF
```

### 3.2 Create Configuration Script

```bash
sudo -u fb2epub cat > /opt/fb2epub/start.sh << 'EOF'
#!/bin/bash
cd /opt/fb2epub
export PORT=8080
export ENVIRONMENT=production
export TEMP_DIR=/opt/fb2epub/temp
export MAX_FILE_SIZE=52428800
./fb2epub >> logs/app.log 2>&1
EOF

chmod +x /opt/fb2epub/start.sh
```

## Step 4: Create Systemd Service

### 4.1 Create Service File

```bash
sudo cat > /etc/systemd/system/fb2epub.service << EOF
[Unit]
Description=FB2 to EPUB Converter Service
After=network.target

[Service]
Type=simple
User=fb2epub
Group=fb2epub
WorkingDirectory=/opt/fb2epub
ExecStart=/opt/fb2epub/fb2epub
Restart=always
RestartSec=10
StandardOutput=append:/opt/fb2epub/logs/app.log
StandardError=append:/opt/fb2epub/logs/app.log
Environment="PORT=8080"
Environment="ENVIRONMENT=production"
Environment="TEMP_DIR=/opt/fb2epub/temp"
Environment="MAX_FILE_SIZE=52428800"

[Install]
WantedBy=multi-user.target
EOF
```

### 4.2 Enable and Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable fb2epub
sudo systemctl start fb2epub
```

### 4.3 Check Service Status

```bash
sudo systemctl status fb2epub
```

### 4.4 View Logs

```bash
# Systemd logs
sudo journalctl -u fb2epub -f

# Application logs
tail -f /opt/fb2epub/logs/app.log
```

## Step 5: Configure Firewall

### 5.1 Allow HTTP/HTTPS (if using Nginx)

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### 5.2 Allow Application Port (if direct access)

```bash
sudo ufw allow 8080/tcp
```

### 5.3 Enable Firewall

```bash
sudo ufw enable
```

## Step 6: Setup Nginx Reverse Proxy (Recommended)

### 6.1 Install Nginx

```bash
sudo apt install -y nginx
```

### 6.2 Create Nginx Configuration

```bash
sudo cat > /etc/nginx/sites-available/fb2epub << EOF
server {
    listen 80;
    server_name your-domain.com;  # Replace with your domain or IP

    client_max_body_size 50M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }
}
EOF
```

### 6.3 Enable Site

```bash
sudo ln -s /etc/nginx/sites-available/fb2epub /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Step 7: Setup SSL with Let's Encrypt (Optional but Recommended)

### 7.1 Install Certbot

```bash
sudo apt install -y certbot python3-certbot-nginx
```

### 7.2 Obtain SSL Certificate

```bash
sudo certbot --nginx -d your-domain.com
```

### 7.3 Auto-renewal (already configured by certbot)

```bash
sudo certbot renew --dry-run
```

## Step 8: Verify Deployment

### 8.1 Test Health Endpoint

```bash
curl http://localhost:8080/health
# or if using Nginx
curl http://your-domain.com/health
```

Expected response:
```json
{"status":"ok","service":"fb2epub"}
```

### 8.2 Test Conversion

```bash
# Upload a test FB2 file
curl -X POST http://your-domain.com/api/v1/convert \
  -F "file=@test.fb2"
```

## Step 9: Maintenance Commands

### Restart Service
```bash
sudo systemctl restart fb2epub
```

### Stop Service
```bash
sudo systemctl stop fb2epub
```

### View Logs
```bash
sudo journalctl -u fb2epub -n 100 -f
```

### Update Application
```bash
cd /opt/fb2epub
sudo -u fb2epub git pull  # if using git
# or upload new binary
sudo systemctl restart fb2epub
```

### Clean Temporary Files
```bash
sudo -u fb2epub find /opt/fb2epub/temp -type f -mtime +7 -delete
```

## Step 10: Monitoring (Optional)

### 10.1 Setup Log Rotation

```bash
sudo cat > /etc/logrotate.d/fb2epub << EOF
/opt/fb2epub/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 fb2epub fb2epub
}
EOF
```

### 10.2 Monitor Disk Space

Add to crontab:
```bash
sudo crontab -e
# Add line:
0 2 * * * find /opt/fb2epub/temp -type f -mtime +1 -delete
```

## Troubleshooting

### Service won't start
```bash
# Check logs
sudo journalctl -u fb2epub -n 50

# Check if port is in use
sudo netstat -tulpn | grep 8080

# Check permissions
ls -la /opt/fb2epub
```

### Permission denied
```bash
sudo chown -R fb2epub:fb2epub /opt/fb2epub
sudo chmod +x /opt/fb2epub/fb2epub
```

### Out of disk space
```bash
# Check disk usage
df -h

# Clean temp files
sudo -u fb2epub find /opt/fb2epub/temp -type f -delete
```

### Nginx 502 Bad Gateway
- Check if service is running: `sudo systemctl status fb2epub`
- Check if service is listening: `sudo netstat -tulpn | grep 8080`
- Check Nginx error logs: `sudo tail -f /var/log/nginx/error.log`

## Security Considerations

1. **Firewall**: Only expose necessary ports
2. **User Permissions**: Run service as non-root user
3. **File Limits**: Set appropriate MAX_FILE_SIZE
4. **Rate Limiting**: Consider adding rate limiting in production
5. **HTTPS**: Always use HTTPS in production
6. **Regular Updates**: Keep system and application updated

## Backup

### Backup Application
```bash
tar -czf fb2epub-backup-$(date +%Y%m%d).tar.gz /opt/fb2epub
```

### Restore
```bash
tar -xzf fb2epub-backup-YYYYMMDD.tar.gz -C /
sudo systemctl restart fb2epub
```

## Quick Deployment Script

Save this as `deploy.sh` and run on your server:

```bash
#!/bin/bash
set -e

# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y curl wget git build-essential nginx

# Install Go
cd /tmp
wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Create user and directories
sudo useradd -m -s /bin/bash fb2epub || true
sudo mkdir -p /opt/fb2epub/{logs,temp}
sudo chown -R fb2epub:fb2epub /opt/fb2epub

# Build and deploy (adjust for your deployment method)
# ... your deployment steps here ...

echo "Deployment complete!"
```

## Support

For issues or questions, check the logs:
- Application logs: `/opt/fb2epub/logs/app.log`
- System logs: `sudo journalctl -u fb2epub`

