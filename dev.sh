#!/bin/bash
# Development script to run backend and frontend concurrently

set -e

echo "🚀 Starting OpenID Connect Server in Development Mode"
echo "======================================================"
echo ""
echo "Backend:  http://localhost:8080"
echo "Frontend: http://localhost:3000"
echo ""

# Function to cleanup background processes on exit
cleanup() {
    echo ""
    echo "🛑 Stopping servers..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup EXIT INT TERM

# Start backend server
echo "Starting backend server..."
cd backend
go run main.go serve &
BACKEND_PID=$!
cd ..

# Wait a moment for backend to start
sleep 2

# Start frontend dev server
echo "Starting frontend dev server..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ Both servers started!"
echo ""
echo "Press Ctrl+C to stop both servers"
echo ""

# Wait for both processes
wait
