#!/bin/bash

echo "Starting Commune in DEVELOPMENT mode..."
echo "=================================="
echo ""

# Check if npm dependencies are installed
echo "Checking npm dependencies..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "node_modules not found. Installing dependencies..."
    npm install
    echo "Dependencies installed successfully!"
elif [ package-lock.json -nt node_modules ]; then
    echo "package-lock.json is newer than node_modules. Updating dependencies..."
    npm install
    echo "Dependencies updated successfully!"
else
    echo "Dependencies are up to date."
fi

# Start Vite dev server in the background
echo "Starting Vite dev server on port 5173..."
npm run dev &
VITE_PID=$!
cd ..

# Wait a bit for Vite to start
sleep 3

# Determine Go bin path
GOPATH_BIN=$(go env GOPATH)/bin
if [ -z "$GOPATH_BIN" ] || [ "$GOPATH_BIN" = "/bin" ]; then
    GOPATH_BIN="$HOME/go/bin"
fi

# Add Go bin to PATH if not already there
if [[ ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    export PATH="$GOPATH_BIN:$PATH"
fi

AIR_BIN="$GOPATH_BIN/air"

# Check if Air is installed
if [ ! -f "$AIR_BIN" ]; then
    echo "Air is not installed. Installing Air..."
    go install github.com/air-verse/air@latest
    echo "Air installed successfully!"
fi

# Generate a random port between 3000 and 9000
BACKEND_PORT=$((3000 + RANDOM % 6001))

# Start Go backend with Air
echo "Starting Go backend with Air on port $BACKEND_PORT..."
cd backend
MODE=development PORT=$BACKEND_PORT "$AIR_BIN" &
GO_PID=$!
cd ..

echo ""
echo "=================================="
echo "Development servers started!"
echo "Frontend (Vite): http://localhost:5173"
echo "Backend (Go): http://localhost:$BACKEND_PORT"
echo "Access the app at: http://localhost:$BACKEND_PORT"
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
