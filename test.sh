#!/bin/bash

# Test script for FB2 to EPUB converter service
# Make sure the service is running on localhost:8080

BASE_URL="http://localhost:8080"

echo "=== Testing FB2 to EPUB Converter Service ==="
echo ""

# Test 1: Health check
echo "1. Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Response: $HEALTH_RESPONSE"
echo ""

# Test 2: Convert FB2 file (if provided)
if [ -z "$1" ]; then
    echo "2. Skipping conversion test (no FB2 file provided)"
    echo "   Usage: ./test.sh path/to/file.fb2"
else
    echo "2. Testing conversion with file: $1"
    
    # Upload file
    CONVERT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/convert" -F "file=@$1")
    echo "Response: $CONVERT_RESPONSE"
    
    # Extract job ID
    JOB_ID=$(echo $CONVERT_RESPONSE | grep -o '"job_id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$JOB_ID" ]; then
        echo "   Error: Could not extract job ID"
        exit 1
    fi
    
    echo "   Job ID: $JOB_ID"
    echo ""
    
    # Wait for processing
    echo "3. Waiting for conversion to complete..."
    for i in {1..30}; do
        sleep 1
        STATUS_RESPONSE=$(curl -s "$BASE_URL/api/v1/status/$JOB_ID")
        STATUS=$(echo $STATUS_RESPONSE | grep -o '"status":"[^"]*' | cut -d'"' -f4)
        
        echo "   Attempt $i: Status = $STATUS"
        
        if [ "$STATUS" = "completed" ]; then
            echo "   Conversion completed!"
            echo ""
            echo "4. Downloading EPUB..."
            curl -s -o "output_${JOB_ID}.epub" "$BASE_URL/api/v1/download/$JOB_ID"
            echo "   EPUB saved as: output_${JOB_ID}.epub"
            break
        elif [ "$STATUS" = "failed" ]; then
            echo "   Conversion failed!"
            echo "   Response: $STATUS_RESPONSE"
            exit 1
        fi
    done
    
    if [ "$STATUS" != "completed" ]; then
        echo "   Timeout waiting for conversion"
        exit 1
    fi
fi

echo ""
echo "=== Tests completed ==="

