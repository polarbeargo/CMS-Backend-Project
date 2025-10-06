# Docker Deployment Guide

This guide provides comprehensive instructions for deploying the CMS Backend Project using Docker and Docker Compose.

## Prerequisites for Docker Deployment

- [Docker](https://docs.docker.com/get-docker/) 20.10 or higher
- [Docker Compose](https://docs.docker.com/compose/install/) 2.0 or higher

## Docker Compose Architecture

The `docker-compose.yml` includes the following services:

- **cms-backend**: Main Go application
- **postgres**: PostgreSQL 15 database with health checks
- **redis**: Redis 7 cache with password protection and LRU eviction
- **redis-commander**: Redis GUI (development profile only)
- **nginx**: Load balancer (production profile only)

## Quick Start with Docker

### Option 1: Basic Development Setup

```bash

cd starter

# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f cms-backend
```

### Option 2: Development with Redis GUI

```bash
# Start with development profile (includes Redis Commander)
docker-compose --profile dev up -d

# Access Redis Commander at http://localhost:8081
# Access CMS API at http://localhost:8080
```

### Option 3: Production with Load Balancer

```bash
# Start with production profile (includes Nginx)
docker-compose --profile production up -d

# Access via Nginx at http://localhost (port 80)
# Access CMS API directly at http://localhost:8080
```

## Docker Environment Configuration

The Docker Compose setup uses optimized environment variables:

```yaml
DB_HOST=postgres
DB_PORT=5432
DB_USER=cms_user
DB_PASSWORD=cms_password
DB_NAME=cms_db
 
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=secure_redis_password
REDIS_DB=0
REDIS_PREFIX=cms_cache:

ENV=production
GIN_MODE=release
```

## Docker Commands Reference

### Starting Services

```bash
# Start all services in background
docker-compose up -d

# Start with specific profile
docker-compose --profile dev up -d
docker-compose --profile production up -d

# Start and rebuild images
docker-compose up -d --build
```

### Monitoring Services

```bash
# View running containers
docker-compose ps

# View service logs
docker-compose logs cms-backend
docker-compose logs postgres
docker-compose logs redis

# Follow logs in real-time
docker-compose logs -f cms-backend
```

### Managing Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (deletes data)
docker-compose down -v

# Restart specific service
docker-compose restart cms-backend

# Scale CMS backend instances
docker-compose up -d --scale cms-backend=3
```

### Database Operations

```bash
# Access PostgreSQL CLI
docker-compose exec postgres psql -U cms_user -d cms_db

# Run database migrations (if needed)
docker-compose exec cms-backend ./migrate.sh up

# Backup database
docker-compose exec postgres pg_dump -U cms_user cms_db > backup.sql
```

### Redis Operations

```bash
# Access Redis CLI
docker-compose exec redis redis-cli -a secure_redis_password

# Check Redis memory usage
docker-compose exec redis redis-cli -a secure_redis_password info memory

# Monitor Redis commands
docker-compose exec redis redis-cli -a secure_redis_password monitor
```

## Service Health Checks

The Docker setup includes health checks for critical services:

### Check Service Health

```bash
# Check all service health
docker-compose ps

# Detailed health status
docker inspect $(docker-compose ps -q postgres) --format='{{json .State.Health}}'
docker inspect $(docker-compose ps -q redis) --format='{{json .State.Health}}'
```

## Data Persistence

The setup includes persistent volumes:

- `postgres_data`: PostgreSQL database files
- `redis_data`: Redis persistence files

### Volume Management

```bash
# List volumes
docker volume ls

# Backup volume data
docker run --rm -v cms-backend-project_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore volume data
docker run --rm -v cms-backend-project_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /data
```

## Development vs Production Profiles

### Development Profile (`--profile dev`)

- Includes Redis Commander GUI at `http://localhost:8081`
- Direct database access on port 5432
- Development-friendly logging
- Hot reload capabilities

### Production Profile (`--profile production`)

- Includes Nginx load balancer
- SSL termination support (requires certificates)
- Production-optimized Redis settings
- Health check monitoring
- Multiple backend instances support

## Troubleshooting Docker Issues

### Common Issues

1. **Port conflicts:**
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   
   # Use different ports
   docker-compose up -d -p 8081:8080
   ```

2. **Database connection errors:**
   ```bash
   # Check PostgreSQL logs
   docker-compose logs postgres
   
   # Verify database is ready
   docker-compose exec postgres pg_isready -U cms_user
   ```

3. **Redis connection issues:**
   ```bash
   # Test Redis connectivity
   docker-compose exec redis redis-cli -a secure_redis_password ping
   
   # Check Redis configuration
   docker-compose exec redis redis-cli -a secure_redis_password config get "*"
   ```

4. **Container resource issues:**
   ```bash
   # Check container resource usage
   docker stats
   
   # Increase memory limits in docker-compose.yml if needed
   ```
## Docker Compose File Structure

The `docker-compose.yml` file is located in the `starter/` directory and includes:

- **Multi-service architecture** with proper service dependencies
- **Environment variable configuration** for all services
- **Volume mounts** for data persistence
- **Health checks** for database and cache services
- **Network isolation** with custom bridge network
- **Profile-based deployment** for different environments
- **Resource limits** and restart policies

For more information about the project structure and API endpoints, see the main [README.md](README.md).