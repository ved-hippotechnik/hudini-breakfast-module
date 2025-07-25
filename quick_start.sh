#!/bin/bash

# Quick Start Script for Hotel Room Grid Dashboard
echo "🏨 Starting Hotel Room Grid Dashboard Setup..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Build and run the main server in background
echo "🔧 Building and starting the backend server..."
go build -o bin/server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to build server"
    exit 1
fi

# Start server in background
./bin/server &
SERVER_PID=$!
echo "✅ Server started with PID $SERVER_PID"

# Wait a moment for server to start
sleep 2

# Create test data
echo "📊 Creating test data..."
go run create_test_data.go
if [ $? -ne 0 ]; then
    echo "❌ Failed to create test data"
    kill $SERVER_PID
    exit 1
fi

echo ""
echo "🎉 Setup complete! Your hotel room grid dashboard is ready."
echo ""
echo "📱 Available Interfaces:"
echo "   • Web Dashboard: Open room-grid-dashboard.html in your browser"
echo "   • API Server: http://localhost:8080"
echo "   • Mobile App: cd mobile && npm start"
echo ""
echo "🏨 Test Data Summary:"
echo "   • Property: Grand Hotel Downtown (PROP001)"
echo "   • Rooms: 50 rooms across 5 floors"
echo "   • Sample guests with various breakfast scenarios"
echo "   • Staff accounts: admin, manager, frontdesk"
echo ""
echo "🔑 Test Credentials:"
echo "   • Username: admin / Password: test"
echo "   • Username: manager / Password: test"
echo "   • Username: frontdesk / Password: test"
echo ""
echo "📋 Room Status Examples:"
echo "   • Rooms 101, 102, 302: Breakfast consumed today"
echo "   • Rooms 103, 201, 401, 501: Breakfast packages pending"
echo "   • Rooms 202, 301: Occupied (no breakfast)"
echo "   • Rooms 105, 205: Maintenance"
echo "   • Room 305: Out of order"
echo ""
echo "🛑 To stop the server: kill $SERVER_PID"
echo "   Or use: pkill -f 'bin/server'"

# Keep script running to show logs
echo ""
echo "📡 Server is running... Press Ctrl+C to stop"
wait $SERVER_PID
