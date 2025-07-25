#!/bin/bash

# Demo Script - Hotel Room Grid Dashboard
echo "🎯 Hotel Room Grid Dashboard Demo"
echo "================================="
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Server is not running. Please start it first with:"
    echo "   ./quick_start.sh"
    echo "   or"
    echo "   go run cmd/server/main.go"
    exit 1
fi

echo "✅ Server is running!"
echo ""

# Test API endpoints
echo "🔍 Testing API endpoints..."
echo ""

echo "1. Getting room grid data:"
curl -s "http://localhost:8080/api/room-grid/PROP001" | jq '.rooms[] | {room_number, status, has_guest, breakfast_package, consumed_today}' | head -20

echo ""
echo "2. Getting specific room details (Room 101):"
curl -s "http://localhost:8080/api/room-grid/PROP001/room/101" | jq '.room'

echo ""
echo "3. Room status summary:"
curl -s "http://localhost:8080/api/room-grid/PROP001" | jq '
{
  total_rooms: (.rooms | length),
  occupied: (.rooms | map(select(.has_guest)) | length),
  breakfast_packages: (.rooms | map(select(.breakfast_package)) | length),
  consumed_today: (.rooms | map(select(.consumed_today)) | length),
  maintenance: (.rooms | map(select(.status == "maintenance")) | length)
}'

echo ""
echo "📱 Access Methods:"
echo "   • Web Dashboard: Open room-grid-dashboard.html"
echo "   • Direct API: http://localhost:8080/api/room-grid/PROP001"
echo "   • Mobile App: cd mobile && npm start"
echo ""
echo "🎨 Visual Features:"
echo "   • Color-coded room status (Green=Vacant, Blue=Breakfast, etc.)"
echo "   • Floor-organized layout"
echo "   • Click/tap rooms for detailed information"
echo "   • Real-time status updates"
echo "   • Search and filter capabilities"
echo ""
echo "Demo complete! 🎉"
