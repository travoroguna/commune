#!/bin/bash

echo "Building Commune for PRODUCTION..."
echo "=================================="
echo ""

# Build frontend
echo "Building frontend..."
cd frontend
npm run build
if [ $? -ne 0 ]; then
    echo "Frontend build failed!"
    exit 1
fi
cd ..

echo "Frontend built successfully!"
echo ""

# Build backend
echo "Building backend..."
cd backend
go build -o commune main.go
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi
cd ..

echo "Backend built successfully!"
echo ""
echo "=================================="
echo "Build complete!"
echo ""
echo "To run in production mode:"
echo "  cd backend"
echo "  MODE=production PORT=8080 ./commune"
echo "=================================="
