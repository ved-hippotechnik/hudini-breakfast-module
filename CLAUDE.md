# Claude Assistant Context - Hudini Breakfast Module

## Project Overview
The Hudini Breakfast Module is a real-time hotel breakfast management system that tracks guest breakfast consumption, provides analytics, and supports multiple hotel properties. It's designed for horizontal scaling and works across all devices.

## Architecture Overview
- **Backend**: Go (Gin framework) REST API with WebSocket support
- **Database**: PostgreSQL (production) / SQLite (development)
- **Cache**: Redis for distributed caching and WebSocket state
- **Frontend**: Multiple interfaces - PWA, React Native mobile, web dashboards
- **Infrastructure**: Docker, Nginx load balancer, horizontal scaling support

## Key Technical Decisions

### Database
- **File**: `internal/database/database.go`
- Supports both PostgreSQL and SQLite based on DATABASE_URL prefix
- PostgreSQL for production (concurrent writes, better performance)
- SQLite for local development (zero configuration)
- Connection pooling configured with sensible defaults

### API Design
- RESTful endpoints under `/api/`
- JWT authentication for protected routes
- Demo endpoints under `/api/demo/` (no auth required)
- WebSocket endpoint at `/ws` for real-time updates
- Health check endpoints: `/health`, `/metrics`, `/ready`, `/live`

### Caching Strategy
- Redis for API response caching
- Distributed WebSocket state via Redis pub/sub
- Cache keys follow pattern: `hudini:<resource>:<identifier>`
- Default TTL: 5 minutes for most cached data

### Security
- JWT tokens with 32+ character secret
- Rate limiting: API (10 req/s), Auth (5 req/min)
- Security headers (CSP, HSTS, XSS protection)
- Input validation on all endpoints
- Bcrypt for password hashing

## Project Structure
```
/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── api/                    # HTTP handlers and routes
│   ├── cache/                  # Redis cache implementation
│   ├── config/                 # Configuration management
│   ├── database/               # Database connection and setup
│   ├── middleware/             # HTTP middleware (auth, security)
│   ├── models/                 # GORM database models
│   ├── services/               # Business logic layer
│   ├── validation/             # Input validation
│   └── websocket/              # WebSocket hub and handlers
├── mobile/                     # React Native mobile app
├── Dockerfile                  # Multi-stage Docker build
├── docker-compose.yml          # Full stack configuration
├── nginx.conf                  # Load balancer config
└── pwa-index.html             # Progressive Web App
```

## Development Guidelines

### Code Style
- Follow Go idioms and conventions
- Use meaningful variable names
- Add error context when wrapping errors
- Keep functions focused and small
- Use interfaces for testability

### Error Handling
- Always check errors immediately
- Wrap errors with context using `fmt.Errorf`
- Log errors at the point of occurrence
- Return appropriate HTTP status codes

### Testing
- Unit tests for services and utilities
- Integration tests for API endpoints
- Use table-driven tests where appropriate
- Mock external dependencies

### Git Workflow
- Commit messages should be descriptive
- Include "why" not just "what"
- Reference issue numbers when applicable
- Keep commits atomic and focused

## Common Tasks

### Running Locally
```bash
# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Run with SQLite (development)
go run cmd/server/main.go

# Run with Docker Compose (full stack)
docker-compose up
```

### Adding New Endpoints
1. Define handler in `internal/api/`
2. Add route in `internal/api/routes.go`
3. Implement business logic in `internal/services/`
4. Add validation in `internal/validation/`
5. Update API documentation

### Database Migrations
- GORM AutoMigrate handles schema updates
- For complex migrations, create migration scripts
- Always backup before running migrations
- Test migrations on staging first

### Debugging WebSocket Issues
1. Check browser console for connection errors
2. Verify `/ws` endpoint is accessible
3. Check Redis pub/sub is working
4. Look for errors in distributed hub logs

## Environment Variables
```
# Database
DATABASE_URL=postgres://user:pass@host/db  # or sqlite://breakfast.db

# Redis
REDIS_URL=redis://:password@host:6379/0

# Security
JWT_SECRET=minimum-32-characters-long-secret

# Server
PORT=8080
SERVER_ID=unique-identifier-for-this-instance
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Feature Flags
ENABLE_ANALYTICS=true
ENABLE_NOTIFICATIONS=true
```

## Performance Considerations
- Database queries use indexes on frequently searched fields
- Pagination implemented for list endpoints
- Caching reduces database load
- WebSocket connections managed efficiently
- Static assets served with proper cache headers

## Known Issues / TODOs
- Kubernetes manifests need to be created
- More comprehensive test coverage needed
- API documentation (OpenAPI/Swagger) pending
- Backup files (*.backup) should be removed
- Consider implementing GraphQL for complex queries

## Troubleshooting

### Database Connection Issues
- Check DATABASE_URL format
- Verify database server is running
- Check firewall/security group rules
- Look for connection pool exhaustion

### Redis Connection Issues
- Verify Redis server is running
- Check authentication credentials
- Monitor Redis memory usage
- Check for network connectivity

### WebSocket Connection Drops
- Check Nginx timeout settings
- Verify Redis pub/sub working
- Look for memory leaks
- Check client-side error handling

## Important Files to Review
1. `internal/database/database.go` - Database configuration
2. `internal/api/routes.go` - API endpoint definitions
3. `internal/services/breakfast.go` - Core business logic
4. `internal/websocket/distributed_hub.go` - Real-time updates
5. `docker-compose.yml` - Full stack setup

## Contact and Resources
- Project documentation: See README.md
- API examples: See test files
- Deployment guide: See SCALABILITY_IMPROVEMENTS.md
- For questions: Check existing issues first

Remember: This is a production system handling real hotel operations. Always test thoroughly and consider the impact of changes on existing data and users.