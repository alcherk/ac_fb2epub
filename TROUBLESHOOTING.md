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

