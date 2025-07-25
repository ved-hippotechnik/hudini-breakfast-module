# 🏨 Hotel Room Grid Dashboard

## Overview

The Hotel Room Grid Dashboard is a comprehensive visual tool for monitoring hotel room statuses, guest information, and breakfast package consumption in real-time. It provides both web-based and mobile interfaces for hotel staff to efficiently manage breakfast services.

## Features

### 🎯 Core Functionality
- **Visual Room Grid**: Color-coded room status display organized by floor
- **Real-time Status Updates**: Live room and breakfast consumption tracking
- **Guest Information**: Detailed guest profiles with breakfast package details
- **Breakfast Management**: Mark and track breakfast consumption
- **Multi-platform Support**: Web dashboard and React Native mobile app

### 🎨 Visual Status Indicators

#### Room Status Colors
- **🟢 Green (Vacant)**: Room is available for guests
- **🟡 Yellow (Occupied)**: Room is occupied but guest has no breakfast package
- **🔵 Blue (Breakfast Package)**: Room is occupied with active breakfast package
- **🔷 Light Blue (Consumed)**: Breakfast has been consumed today
- **🔴 Red (Maintenance)**: Room is under maintenance
- **⚫ Gray (Out of Order)**: Room is out of service

#### Additional Indicators
- **Blue Dot**: Guest has breakfast package (not consumed)
- **Teal Dot**: Breakfast consumed today

### 📊 Dashboard Statistics
- **Total Rooms**: Overall room count
- **Occupied Rooms**: Currently occupied rooms
- **Breakfast Packages**: Active breakfast packages
- **Consumed Today**: Breakfasts consumed today
- **Pending Consumption**: Outstanding breakfast packages

## Interface Components

### Web Dashboard (`room-grid-dashboard.html`)

#### Header Section
- Hotel name and branding
- Date picker for viewing different dates
- PMS sync and refresh buttons

#### Status Legend
- Visual guide showing all room status colors
- Clear explanations for each status type

#### Statistics Overview
- Real-time counts and metrics
- Color-coded statistics cards

#### Room Grid
- Floor-organized room layout
- Clickable room cards with detailed information
- Responsive grid that adapts to screen size

#### Room Details Modal
- Comprehensive guest information
- Check-in/check-out dates
- Breakfast package details
- Consumption history
- Action buttons for staff operations

### Mobile App (`mobile/app/(tabs)/room-grid.tsx`)

#### Features
- Touch-optimized room grid
- Pull-to-refresh functionality
- Search and filtering capabilities
- Modal-based room details
- Quick breakfast marking actions

#### Mobile-Specific Enhancements
- Responsive grid layout (4 rooms per row)
- Touch-friendly buttons and interactions
- Optimized loading states
- Native mobile navigation

## Technical Implementation

### Backend API Endpoints

```
GET /api/room-grid/:propertyId
- Retrieves room grid data for a specific property
- Optional date parameter for historical views

GET /api/room-grid/:propertyId/room/:roomNumber
- Gets detailed information for a specific room

POST /api/room-grid/:propertyId/room/:roomNumber/consume
- Marks breakfast as consumed for a room
- Requires staff authentication

POST /api/room-grid/:propertyId/sync-pms
- Synchronizes guest data from Property Management System
```

### Data Models

#### Room Status Structure
```typescript
interface RoomBreakfastStatus {
  property_id: string;
  room_number: string;
  floor: number;
  room_type: string;
  status: string;
  has_guest: boolean;
  guest_name: string;
  breakfast_package: boolean;
  breakfast_count: number;
  consumed_today: boolean;
  consumed_at?: string;
  consumed_by: string;
  check_in_date?: string;
  check_out_date?: string;
}
```

## Usage Instructions

### For Hotel Staff

#### Daily Operations
1. **Morning Setup**
   - Open the room grid dashboard
   - Verify current date is selected
   - Review overnight status changes
   - Sync with PMS if needed

2. **Breakfast Service**
   - Monitor rooms with breakfast packages (blue indicators)
   - Click on room cards to view guest details
   - Mark breakfasts as consumed when served
   - Track consumption progress throughout service hours

3. **Status Monitoring**
   - Check for maintenance/out-of-order rooms
   - Monitor occupancy levels
   - Review pending breakfast services

#### Mobile Operations
1. **On-the-Go Monitoring**
   - Use mobile app for real-time updates
   - Filter rooms by status type
   - Search for specific rooms or guests
   - Mark breakfast consumption directly from mobile

### For Management

#### Daily Reports
- Review consumption statistics
- Monitor staff performance
- Track breakfast package utilization
- Analyze occupancy patterns

#### Historical Analysis
- Use date picker to review past dates
- Compare consumption patterns
- Identify trends and opportunities

## Setup and Installation

### Prerequisites
- Go 1.21+ (backend)
- Node.js 18+ (mobile app)
- SQLite database
- Modern web browser

### Backend Setup
```bash
# Start the Go server
go run cmd/server/main.go

# Create test data (optional)
go run create_test_data.go
```

### Web Dashboard
```bash
# Open room-grid-dashboard.html in a web browser
# Ensure backend server is running on localhost:8080
```

### Mobile App
```bash
cd mobile
npm install
npm start
# Follow Expo CLI instructions to run on device/simulator
```

## Configuration

### API Configuration
- Update `API_BASE_URL` in dashboard JavaScript
- Configure `PROPERTY_ID` for your hotel
- Set authentication tokens for staff access

### Customization
- Modify colors in CSS/StyleSheet for branding
- Adjust room grid layout for different screen sizes
- Add custom fields for property-specific needs

## Troubleshooting

### Common Issues

#### Data Not Loading
- Verify backend server is running
- Check browser console for API errors
- Ensure correct property ID configuration

#### Room Status Not Updating
- Check database connections
- Verify PMS integration is working
- Refresh data manually if needed

#### Mobile App Issues
- Ensure correct API endpoint configuration
- Check network connectivity
- Verify authentication tokens

### Support
For technical support or feature requests, please refer to the main project documentation or contact the development team.

## Future Enhancements

### Planned Features
- **Real-time Notifications**: Push notifications for status changes
- **Advanced Analytics**: Detailed reporting and business intelligence
- **Integration Improvements**: Enhanced PMS and payment system integration
- **Multi-language Support**: Internationalization for global properties
- **Automated Workflows**: Smart automation for common operations

### Customization Options
- **Theming**: Custom color schemes and branding
- **Layout Options**: Different grid layouts and viewing modes
- **Custom Fields**: Property-specific data fields and requirements
- **Advanced Filters**: More sophisticated filtering and search options
