# 🍳 Hudini Breakfast Module

A comprehensive hotel breakfast consumption tracking system built with Go backend and React Native mobile app, featuring real-time room status monitoring, guest management, and multiple payment method support.

## 🚀 Features

### 🏨 Core Functionality
- **Real-time Room Status Tracking** - Monitor breakfast consumption across all hotel rooms
- **Guest Management System** - Complete guest lifecycle management with breakfast packages
- **Multi-Payment Support** - Room charge, OHIP, cash, and complimentary options
- **Staff Authentication** - Role-based access control (Admin, Manager, Staff)
- **Consumption History** - Detailed tracking of all breakfast activities
- **Daily Reports & Analytics** - Comprehensive reporting and business intelligence

### 🔧 Technical Features
- **RESTful API** - Complete backend API with authentication
- **Web Interface** - Full-featured web application
- **React Native App** - Mobile application for on-the-go management
- **SQLite Database** - Reliable data persistence
- **CORS Support** - Cross-origin resource sharing enabled
- **JWT Authentication** - Secure token-based authentication

## 🛠️ Technology Stack

### Backend
- **Go 1.21+** - High-performance backend server
- **Gin Framework** - Fast HTTP web framework
- **GORM** - Go ORM for database operations
- **SQLite** - Lightweight database
- **JWT** - JSON Web Tokens for authentication
- **bcrypt** - Password hashing

### Frontend
- **React Native** - Mobile application framework
- **Expo** - React Native development platform
- **HTML/CSS/JavaScript** - Web interface
- **Modern CSS Grid** - Responsive design

### Database Schema
- **Staff Management** - User roles and authentication
- **Room Management** - Hotel room status and types
- **Guest Management** - Guest information and breakfast packages
- **Consumption Tracking** - Daily breakfast consumption records
- **OHIP Integration** - Healthcare payment processing

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Node.js 18+ (for mobile app)
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/ved@hippotechnik.com/hudini-breakfast-module.git
   cd hudini-breakfast-module
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install Go dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

4. **Create dummy data (optional)**
   ```bash
   go run create_dummy_data.go
   ```

5. **Start the backend server**
   ```bash
   go run cmd/server/main.go
   ```
   Server will start on `http://localhost:3001`

6. **Start the web interface**
   ```bash
   python3 -m http.server 8000
   ```
   Web interface available at `http://localhost:8000/complete-app.html`

### Mobile App Setup

1. **Navigate to mobile directory**
   ```bash
   cd mobile
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Start the app**
   ```bash
   npm start
   # or
   npx expo start
   ```

## 📱 Usage

### Web Interface
Access the complete web application at `http://localhost:8000/complete-app.html`

#### Test Credentials (after running dummy data script):
- **Admin**: `admin@hotel.com` / `password123`
- **Manager**: `manager@hotel.com` / `password123`
- **Staff**: `staff@hotel.com` / `password123`

### API Endpoints

#### Authentication
- `POST /api/auth/register` - Register new staff member
- `POST /api/auth/login` - Staff login
- `GET /api/auth/me` - Get current user profile

#### Room Management
- `GET /api/rooms/breakfast-status` - Get room breakfast status
- `POST /api/rooms/{room_number}/consume` - Mark breakfast consumed

#### Guest Management
- `GET /api/guests` - Get all guests
- `POST /api/guests` - Create new guest
- `PUT /api/guests/{id}` - Update guest information

#### Reports & Analytics
- `GET /api/consumption/history` - Get consumption history
- `GET /api/reports/daily` - Get daily reports
- `GET /api/analytics` - Get analytics data

#### Health Check
- `GET /health` - Server health status

## 🏗️ Project Structure

```
hudini-breakfast-module/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # API handlers and routes
│   ├── config/                  # Configuration management
│   ├── database/                # Database initialization
│   ├── models/                  # Data models
│   └── services/                # Business logic
├── mobile/                      # React Native mobile app
│   ├── app/                     # App screens and navigation
│   ├── src/                     # Source code
│   └── package.json
├── web-interfaces/              # Web frontend files
├── bin/                         # Compiled binaries
├── create_dummy_data.go         # Dummy data generation
├── .env.example                 # Environment variables template
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## 🎯 Key Features Explained

### Room Status Tracking
- Real-time visual grid of all hotel rooms
- Color-coded status indicators (Available, Occupied, Maintenance)
- Breakfast package status for each guest
- One-click consumption marking

### Guest Management
- Complete guest information storage
- Check-in/check-out date tracking
- Breakfast package assignment
- OHIP number support for healthcare integration

### Payment Processing
- **Room Charge** - Bill to guest room
- **OHIP** - Healthcare payment processing
- **Cash** - Direct cash payment
- **Complimentary** - Free breakfast offerings

### Staff Management
- Role-based access control
- Secure authentication system
- Activity tracking and audit trails
- Multi-property support

### Reporting & Analytics
- Daily consumption reports
- Revenue tracking
- Consumption rate analysis
- Historical data trends
- Export capabilities

## 🔧 Configuration

### Environment Variables (.env)
```bash
# Server Configuration
PORT=3001
GIN_MODE=debug

# Database Configuration
DATABASE_URL=sqlite://breakfast.db

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# OHIP Integration
OHIP_BASE_URL=https://api.oraclehospitality.com
OHIP_CLIENT_ID=your-ohip-client-id
OHIP_CLIENT_SECRET=your-ohip-client-secret
```

## 🧪 Testing

### Sample Data
Run the dummy data script to populate the system with realistic test data:
```bash
go run create_dummy_data.go
```

This creates:
- 50 hotel rooms across 5 floors
- 19 active guests with breakfast packages
- 54 consumption records over 7 days
- 5 staff members with different roles

### API Testing
Use the built-in API testing features in the web interface or tools like Postman to test endpoints.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Email: ved@hippotechnik.com
- Create an issue in this repository

## 🙏 Acknowledgments

- Built with Go and Gin framework
- React Native for mobile development
- GORM for database operations
- Modern web technologies for responsive UI

---

**Hudini Breakfast Module** - Streamlining hotel breakfast service management with modern technology.
