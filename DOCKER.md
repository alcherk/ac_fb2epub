# Docker Deployment Guide

Quick reference for Docker deployment of FB2 to EPUB converter.

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start the service
docker-compose up -d

# Access the web UI at http://localhost:3080

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Using Docker

```bash
# Build image
docker build -t fb2epub .

# Run container
docker run -d \
  --name fb2epub \
  -p 3080:8080 \
  -v $(pwd)/temp:/app/temp \
  fb2epub

# View logs
docker logs -f fb2epub

# Stop container
docker stop fb2epub
docker rm fb2epub
```

## Configuration

### Environment Variables

You can override default settings via environment variables in `docker-compose.yml`:

```yaml
environment:
  - PORT=8080
  - ENVIRONMENT=production
  - TEMP_DIR=/app/temp
  - MAX_FILE_SIZE=52428800
  - CLEANUP_TRIGGER_COUNT=10
```

Or when using `docker run`:

```bash
docker run -d \
  --name fb2epub \
  -p 3080:8080 \
  -e PORT=8080 \
  -e ENVIRONMENT=production \
  -e MAX_FILE_SIZE=104857600 \
  fb2epub
```

### Volumes

The `temp` directory is mounted as a volume to persist temporary files:

```yaml
volumes:
  - ./temp:/app/temp
```

## Building

### Build locally

```bash
docker build -t fb2epub .
```

### Build with specific tag

```bash
docker build -t fb2epub:latest -t fb2epub:v1.0.0 .
```

### Multi-stage build

The Dockerfile uses a multi-stage build:
1. **Builder stage**: Compiles the Go application
2. **Final stage**: Minimal Alpine image with just the binary

This results in a small (~20MB) final image.

## Health Checks

The container includes a health check:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

Check health status:

```bash
docker ps
# Look for "healthy" status

docker inspect --format='{{.State.Health.Status}}' fb2epub
```

## Networking

### Expose on different port

```bash
docker run -d -p 3000:8080 fb2epub
```

**Note:** Default Docker port is now 3080 (host) mapped to 8080 (container).

### Use host network (Linux only)

```bash
docker run -d --network host fb2epub
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs fb2epub

# Run interactively to see errors
docker run -it fb2epub
```

### Permission issues

```bash
# Fix temp directory permissions
sudo chown -R 1000:1000 ./temp
```

### Port already in use

```bash
# Change port mapping (default is 3080:8080)
docker run -d -p 3081:8080 fb2epub
```

### Out of disk space

```bash
# Clean up Docker
docker system prune -a

# Remove old images
docker image prune -a
```

## Production Deployment

### With Docker Compose

1. Create production `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  fb2epub:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: fb2epub
    ports:
      - "127.0.0.1:3080:8080"  # Only bind to localhost
    environment:
      - PORT=8080
      - ENVIRONMENT=production
      - TEMP_DIR=/app/temp
      - MAX_FILE_SIZE=52428800
    volumes:
      - ./temp:/app/temp
      - ./logs:/app/logs
    restart: always
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

2. Run with production config:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Behind Nginx

See `DEPLOYMENT.md` for Nginx configuration. The container should bind to `127.0.0.1:3080` (host) or use an internal Docker network (container port 8080).

## Updating

```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose up -d --build

# Or with Docker
docker build -t fb2epub .
docker stop fb2epub
docker rm fb2epub
docker run -d --name fb2epub -p 3080:8080 -v $(pwd)/temp:/app/temp fb2epub
```

## Monitoring

### View logs

```bash
# Docker Compose
docker-compose logs -f fb2epub

# Docker
docker logs -f fb2epub

# Last 100 lines
docker logs --tail 100 fb2epub
```

### Resource usage

```bash
docker stats fb2epub
```

### Container info

```bash
docker inspect fb2epub
```

## Security

1. **Non-root user**: Container runs as `appuser` (UID 1000)
2. **Minimal base image**: Alpine Linux (~5MB)
3. **No unnecessary packages**: Only ca-certificates and tzdata
4. **Read-only root filesystem** (optional):

```yaml
services:
  fb2epub:
    read_only: true
    tmpfs:
      - /tmp
      - /app/temp
```

## Best Practices

1. Use Docker Compose for easier management
2. Set resource limits:

```yaml
services:
  fb2epub:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

3. Use volumes for persistent data
4. Enable health checks
5. Use restart policies: `restart: unless-stopped`
6. Keep images updated
7. Scan images for vulnerabilities:

```bash
docker scan fb2epub
```

