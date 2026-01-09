#!/bin/bash

# Demo mode startup script for Open Exposure Management
# This script starts both the backend and frontend in demo mode

echo "🔓 Starting Open Exposure Management in DEMO MODE"
echo "⚠️  WARNING: Authentication is DISABLED for demonstration purposes"
echo "⚠️  This is NOT secure for production use!"
echo ""

# Set demo mode environment variable for backend
export DEMO_MODE=true

# Start backend in background
echo "🚀 Starting backend server..."
cd api
go run ./cmd/server &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Start frontend in background
echo "🌐 Starting frontend..."
cd ../ui
npm run dev &
FRONTEND_PID=$!

echo ""
echo "✅ Demo mode ready!"
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend: http://localhost:8080"
echo ""
echo "🛑 Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $FRONTEND_PID 2>/dev/null
    kill $BACKEND_PID 2>/dev/null
    echo "✅ Services stopped"
    exit 0
}

# Set trap for cleanup
trap cleanup INT TERM

# Wait for background processes
wait