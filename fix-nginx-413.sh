#!/bin/bash
# Script to fix 413 error by updating nginx configuration
# Run this on your remote server

echo "Checking nginx status..."
if systemctl is-active --quiet nginx; then
    echo "✓ Nginx is running"
else
    echo "✗ Nginx is not running. This script is for nginx configuration."
    exit 1
fi

echo ""
echo "Current nginx client_max_body_size settings:"
sudo nginx -T 2>/dev/null | grep -i "client_max_body_size" || echo "No client_max_body_size found"

echo ""
echo "Checking nginx configuration files..."
CONFIG_FILES=(
    "/etc/nginx/nginx.conf"
    "/etc/nginx/sites-available/default"
    "/etc/nginx/sites-available/fb2epub"
    "/etc/nginx/conf.d/default.conf"
)

for file in "${CONFIG_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "Found: $file"
        if grep -q "client_max_body_size" "$file"; then
            echo "  Current setting: $(grep "client_max_body_size" "$file")"
        else
            echo "  No client_max_body_size setting found"
        fi
    fi
done

echo ""
echo "To fix the 413 error, you need to add or update client_max_body_size:"
echo ""
echo "1. Edit the nginx config file:"
echo "   sudo nano /etc/nginx/nginx.conf"
echo ""
echo "2. Add or update in the http block:"
echo "   http {"
echo "       client_max_body_size 100M;"
echo "       ..."
echo "   }"
echo ""
echo "3. Or add in the server block:"
echo "   server {"
echo "       client_max_body_size 100M;"
echo "       ..."
echo "   }"
echo ""
echo "4. Test and restart:"
echo "   sudo nginx -t"
echo "   sudo systemctl restart nginx"

