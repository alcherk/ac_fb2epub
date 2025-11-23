#!/bin/bash
# Script to fix nginx configuration for fb2epub.z0mi.cloud
# Run this on your remote server

echo "=== Fixing nginx configuration for fb2epub.z0mi.cloud ==="
echo ""

# Find the config file
CONFIG_FILE="/etc/nginx/sites-available/fb2epub"
if [ ! -f "$CONFIG_FILE" ]; then
    CONFIG_FILE="/etc/nginx/sites-available/default"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Could not find nginx config file"
    echo "Please manually edit the config file for fb2epub.z0mi.cloud"
    exit 1
fi

echo "Found config file: $CONFIG_FILE"
echo ""

# Check if client_max_body_size is in the fb2epub server block
if grep -A 20 "server_name fb2epub.z0mi.cloud" "$CONFIG_FILE" | grep -q "client_max_body_size"; then
    echo "✓ client_max_body_size already found in fb2epub server block"
    grep -A 20 "server_name fb2epub.z0mi.cloud" "$CONFIG_FILE" | grep "client_max_body_size"
else
    echo "✗ client_max_body_size NOT found in fb2epub server block"
    echo ""
    echo "You need to add it manually. Here's what to do:"
    echo ""
    echo "1. Edit the config file:"
    echo "   sudo nano $CONFIG_FILE"
    echo ""
    echo "2. Find the server block for fb2epub.z0mi.cloud:"
    echo "   server {"
    echo "       server_name fb2epub.z0mi.cloud;"
    echo ""
    echo "3. Add these lines right after server_name:"
    echo "       client_max_body_size 100M;"
    echo "       client_body_buffer_size 128k;"
    echo ""
    echo "4. Also add to the location / block:"
    echo "       location / {"
    echo "           proxy_request_buffering off;"
    echo "           proxy_buffering off;"
    echo "           ..."
    echo ""
    echo "5. Test and restart:"
    echo "   sudo nginx -t"
    echo "   sudo systemctl restart nginx"
fi

echo ""
echo "Current fb2epub server block:"
grep -A 15 "server_name fb2epub.z0mi.cloud" "$CONFIG_FILE" || echo "Not found"

