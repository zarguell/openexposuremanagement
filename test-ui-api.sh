#!/bin/bash

echo "🧪 Testing UI API Integration"
echo "================================"
echo ""

# Test 1: HTML page loads
echo "1️⃣ Testing HTML page load..."
HTML=$(curl -s http://localhost:80/)
if echo "$HTML" | grep -q "Open Exposure Management"; then
    echo "✅ HTML page loads correctly"
else
    echo "❌ HTML page failed to load"
fi
echo ""

# Test 2: API through proxy (what the browser should use)
echo "2️⃣ Testing API through nginx proxy (/api/v1/...)..."
DASHBOARD_PROXY=$(curl -s http://localhost:80/api/v1/dashboard)
if echo "$DASHBOARD_PROXY" | grep -q "Assets"; then
    echo "✅ API proxy works correctly"
    echo "   Response: $DASHBOARD_PROXY" | head -c 200
else
    echo "❌ API proxy failed"
    echo "   Response: $DASHBOARD_PROXY"
fi
echo ""

# Test 3: API directly on backend port
echo "3️⃣ Testing API directly on backend (port 8080)..."
DASHBOARD_DIRECT=$(curl -s http://localhost:8080/api/v1/dashboard)
if echo "$DASHBOARD_DIRECT" | grep -q "Assets"; then
    echo "✅ Backend API works correctly"
    echo "   Response: $DASHBOARD_DIRECT" | head -c 200
else
    echo "❌ Backend API failed"
    echo "   Response: $DASHBOARD_DIRECT"
fi
echo ""

# Test 4: Check nginx access logs
echo "4️⃣ Checking recent nginx access logs..."
docker logs oem-ui 2>&1 | grep "GET /" | tail -5
echo ""

echo "================================"
echo "Test complete!"
