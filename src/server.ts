import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import dotenv from 'dotenv';
import { createServer } from 'http';
import { DatabaseManager } from './database/DatabaseManager';
import { Logger } from './utils/Logger';
import { routes } from './routes';
import { errorHandler } from './middleware/errorHandler';
import { requestLogger } from './middleware/requestLogger';
import { rateLimiter } from './middleware/rateLimiter';
import { HealthChecker } from './services/HealthChecker';

// Load environment variables
dotenv.config();

const app = express();
const server = createServer(app);
const PORT = process.env.PORT || 3001;

// Initialize logger
const logger = Logger.getInstance();

// Security middleware
app.use(helmet());
app.use(cors({
  origin: process.env.CLIENT_URL || 'http://localhost:3000',
  credentials: true
}));

// Request parsing middleware
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Logging and rate limiting
app.use(requestLogger);
app.use(rateLimiter);

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: process.env.npm_package_version || '1.0.0'
  });
});

// API routes
app.use(`/api/${process.env.API_VERSION || 'v1'}`, routes);

// Error handling
app.use(errorHandler);

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Route not found',
    message: `Cannot ${req.method} ${req.originalUrl}`
  });
});

// Graceful shutdown
process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);

async function gracefulShutdown(signal: string) {
  logger.info(`Received ${signal}. Starting graceful shutdown...`);
  
  server.close(() => {
    logger.info('HTTP server closed');
    
    // Close database connections
    DatabaseManager.getInstance().closePool()
      .then(() => {
        logger.info('Database connections closed');
        process.exit(0);
      })
      .catch((error) => {
        logger.error('Error closing database connections:', error);
        process.exit(1);
      });
  });
}

// Start server
async function startServer() {
  try {
    // Initialize database
    await DatabaseManager.getInstance().initialize();
    logger.info('Database initialized successfully');

    // Start health checker
    HealthChecker.getInstance().start();
    
    server.listen(PORT, () => {
      logger.info(`🚀 Breakfast Module server running on port ${PORT}`);
      logger.info(`🏥 OHIP Integration enabled`);
      logger.info(`📊 Health check available at http://localhost:${PORT}/health`);
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    process.exit(1);
  }
}

startServer();

export { app, server };
