# Troubleshooting Guide

## macOS: LC_UUID Error

If you encounter the error:
```
dyld: missing LC_UUID load command
Abort trap: 6
```

This is a known macOS issue that can affect both compiled binaries and `go run`. It's typically caused by:
- Corrupted Go toolchain cache
- macOS System Integrity Protection (SIP) issues
- Go installation problems

### Solution 1: Use Docker (Recommended)

Docker completely avoids this issue and provides a consistent environment:

```bash
# Start Docker Desktop first, then:
docker-compose up -d

# Access at http://localhost:8080
```

### Solution 2: Fix Go Installation

1. **Clear all Go caches**:
   ```bash
   go clean -cache -modcache -testcache
   rm -rf ~/Library/Caches/go-build
   ```

2. **Reinstall Go** (if using Homebrew):
   ```bash
   brew uninstall go
   brew install go
   ```

3. **Or download fresh Go from golang.org**:
   ```bash
   # Remove old installation
   sudo rm -rf /usr/local/go
   
   # Download and install latest
   # Visit https://golang.org/dl/
   ```

4. **Rebuild**:
   ```bash
   rm -f fb2epub
   go build -trimpath -o fb2epub
   ```

### Solution 3: Use Go Install

Sometimes `go install` works when `go build` doesn't:

```bash
go install .
~/go/bin/fb2epub
```

### Solution 4: Code Sign the Binary

```bash
codesign --force --deep --sign - fb2epub
./fb2epub
```

### Solution 5: Run in Different Location

Sometimes the issue is path-related:

```bash
cp fb2epub /tmp/fb2epub
/tmp/fb2epub
```

## Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
export PORT=8081
./fb2epub
```

## Permission Denied

```bash
# Make executable
chmod +x fb2epub

# Or create temp directory
mkdir -p /tmp/fb2epub
chmod 755 /tmp/fb2epub
```

## Web UI Not Loading

1. **Check if web files exist**:
   ```bash
   ls -la web/index.html web/static/
   ```

2. **Check server logs** for file path errors

3. **Verify the service is running**:
   ```bash
   curl http://localhost:8080/health
   ```

## File Upload Fails

1. **Check file size limit**:
   ```bash
   export MAX_FILE_SIZE=104857600  # 100MB

### 413 Request Entity Too Large (HTML error)

If you get a 413 error with HTML response (not JSON), it's likely coming from a reverse proxy (nginx) or the HTTP server itself.

**Diagnosis:**
- If you see HTML error page: The error is from nginx/reverse proxy
- If you see JSON error: The error is from the application
- Check application logs: `docker logs fb2epub | grep "Max file size"` should show 100MB

**Solutions:**

1. **Check if nginx is running:**
   ```bash
   sudo systemctl status nginx
   # or
   ps aux | grep nginx
   ```

2. **Find nginx configuration files:**
   ```bash
   # Check main nginx config
   sudo cat /etc/nginx/nginx.conf | grep -A 5 "client_max_body_size"
   
   # Check site-specific configs
   sudo ls -la /etc/nginx/sites-available/
   sudo ls -la /etc/nginx/sites-enabled/
   
   # Check for Docker-related configs
   sudo grep -r "8080\|3080\|fb2epub" /etc/nginx/
   ```

3. **Update nginx configuration:**
   
   **Option A: Update main nginx.conf (affects all sites):**
   ```bash
   sudo nano /etc/nginx/nginx.conf
   
   # Add or update in http block:
   http {
       client_max_body_size 100M;  # Add this line
       ...
   }
   ```

   **Option B: Update site-specific config:**
   ```bash
   # Find the config file for your site
   sudo nano /etc/nginx/sites-available/fb2epub
   # or
   sudo nano /etc/nginx/sites-available/default
   
   # Add or update in server block:
   server {
       client_max_body_size 100M;  # Must match or exceed MAX_FILE_SIZE
       ...
   }
   ```

4. **Test and restart nginx:**
   ```bash
   # Test configuration
   sudo nginx -t
   
   # If test passes, restart nginx
   sudo systemctl restart nginx
   
   # Or reload without downtime
   sudo systemctl reload nginx
   ```

5. **Verify the fix:**
   ```bash
   # Check nginx config
   sudo nginx -T | grep client_max_body_size
   
   # Should show: client_max_body_size 100M;
   ```

6. **If using Docker with nginx in front:**
   ```bash
   # Check docker-compose.yml for nginx service
   # Or check if nginx is running as a separate container
   docker ps | grep nginx
   ```

7. **Check nginx error logs:**
   ```bash
   sudo tail -f /var/log/nginx/error.log
   # Upload a file and watch for 413 errors
   ```

**Important Notes:**
- `client_max_body_size` must be set in the `http` block or `server` block
- **If you have multiple server blocks, each one needs its own setting**
- Setting it in `location` block also works but is less common
- The value must match or exceed your `MAX_FILE_SIZE` (100MB = 104857600 bytes)
- After changing nginx config, always run `sudo nginx -t` before restarting

**Common Issue: Multiple Server Blocks**

If you have multiple server blocks (e.g., one for main site, one for fb2epub subdomain), 
you need to set `client_max_body_size` in EACH server block:

```nginx
# Main site
server {
    server_name z0mi.cloud;
    client_max_body_size 100M;  # ← Needed here
    ...
}

# FB2EPUB subdomain
server {
    server_name fb2epub.z0mi.cloud;
    client_max_body_size 100M;  # ← Also needed here!
    
    location / {
        proxy_pass http://127.0.0.1:3080;
        # Optional: Also add these for better handling
        proxy_request_buffering off;
        proxy_buffering off;
        ...
    }
}
```

**If still getting 413 after setting client_max_body_size:**

1. **Check nginx error logs:**
   ```bash
   sudo tail -f /var/log/nginx/error.log
   # Try uploading and watch for errors
   ```

2. **Verify the setting is in the correct server block:**
   ```bash
   sudo nginx -T | grep -B 5 "fb2epub.z0mi.cloud" | grep -A 10 "client_max_body_size"
   ```

3. **Try adding to location block as well:**
   ```nginx
   location / {
       client_max_body_size 100M;  # Override if needed
       proxy_pass http://127.0.0.1:3080;
       ...
   }
   ```

4. **Disable proxy buffering (helps with large uploads):**
   ```nginx
   location / {
       proxy_request_buffering off;
       proxy_buffering off;
       proxy_pass http://127.0.0.1:3080;
       ...
   }
   ```
   ```

2. **Verify file type**: Must be `.fb2` or `.xml`

3. **Check temp directory permissions**:
   ```bash
   ls -la /tmp/fb2epub
   ```

## Conversion Fails

1. **Check if FB2 file is valid XML**:
   ```bash
   xmllint --noout your-file.fb2
   ```

2. **View server logs** for detailed error messages

3. **Check job status**:
   ```bash
   curl http://localhost:8080/api/v1/status/{job_id}
   ```

## Docker Issues

### Container won't start
```bash
# Check logs
docker logs fb2epub

# Run interactively
docker run -it fb2epub
```

### Permission issues
```bash
# Fix temp directory
sudo chown -R 1000:1000 ./temp
```

### Port conflicts
```bash
# Change port in docker-compose.yml
ports:
  - "8081:8080"
```

## General Debugging

1. **Enable verbose logging**: Set `ENVIRONMENT=development`

2. **Check Go version**:
   ```bash
   go version
   # Should be 1.21 or later
   ```

3. **Verify dependencies**:
   ```bash
   go mod download
   go mod verify
   ```

4. **Clean build**:
   ```bash
   make clean
   go mod tidy
   make build
   ```

## Getting Help

1. Check the logs:
   - Application: `./fb2epub` output
   - Docker: `docker logs fb2epub`

2. Test the health endpoint:
   ```bash
   curl http://localhost:8080/health
   ```

3. Verify the service is running:
   ```bash
   ps aux | grep fb2epub
   ```

