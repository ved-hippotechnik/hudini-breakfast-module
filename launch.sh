#!/bin/bash

# Hudini Breakfast Module - Launch Script
echo "🏨 Launching Hudini Breakfast Module..."

# Set Go modules
export GO111MODULE=on

# Check if server is already running
if pgrep -f "bin/server" > /dev/null; then
    echo "✅ Server already running on port 8080"
else
    echo "🚀 Starting Go server..."
    ./bin/server &
    sleep 3
fi

# Open the main dashboards
echo "📊 Opening Enhanced Room Grid Dashboard..."
open "enhanced-dashboard.html"

echo "📈 Opening Advanced Analytics Dashboard..."
open "analytics-dashboard.html"

# Start mobile app if requested
if [ "$1" = "--mobile" ]; then
    echo "📱 Starting mobile app..."
    cd mobile && npm start &
fi

echo ""
echo "🎉 Hudini Breakfast Module is now running!"
echo ""
echo "Available dashboards:"
echo "  • Enhanced Room Grid: file://$(pwd)/enhanced-dashboard.html"
echo "  • Advanced Analytics: file://$(pwd)/analytics-dashboard.html"
echo ""
echo "API Server: http://localhost:8080"
echo "Health Check: http://localhost:8080/health"
echo ""
echo "Analytics Endpoints:"
echo "  • Basic Analytics: http://localhost:8080/api/analytics"
echo "  • Advanced Analytics: http://localhost:8080/api/analytics/advanced"
echo "  • Real-time Metrics: http://localhost:8080/api/analytics/realtime"
echo "  • Predictive Insights: http://localhost:8080/api/analytics/predictive"
echo "  • Business Intelligence: http://localhost:8080/api/analytics/business-intelligence"
echo ""
echo "To stop: pkill -f 'bin/server'"
echo "For mobile app: Add --mobile flag"
