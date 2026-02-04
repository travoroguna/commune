#!/bin/bash

echo "Starting Commune in DEVELOPMENT mode..."
echo "=================================="
echo ""

# Start Vite dev server in the background
echo "Starting Vite dev server on port 5173..."
cd frontend
npm run dev &
VITE_PID=$!
cd ..

# Wait a bit for Vite to start
sleep 3

# Start Go backend
echo "Starting Go backend on port 3000..."
cd backend
MODE=development PORT=3000 go run main.go &
GO_PID=$!
cd ..

echo ""
echo "=================================="
echo "Development servers started!"
echo "Frontend (Vite): http://localhost:5173"
echo "Backend (Go): http://localhost:3000"
echo "Access the app at: http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all servers"
echo "=================================="

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping servers..."
    kill $VITE_PID 2>/dev/null
    kill $GO_PID 2>/dev/null
    echo "All servers stopped."
    exit 0
}

# Trap Ctrl+C
trap cleanup INT

# Wait for processes
wait
