# Database and Redis Setup Guide

This guide provides detailed instructions for setting up PostgreSQL database and Redis cache for the CMS Backend Project.

## Prerequisites

Before proceeding, ensure you have the following installed:
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/) (For external caching)
- [Homebrew](https://brew.sh/) (for macOS users)

## PostgreSQL Database Setup

You can set up your PostgreSQL database using either the PostgreSQL CLI or pgAdmin 4.

### Option 1: Using PostgreSQL CLI

1. **Start PostgreSQL service using the following commands:**

    ```bash
    brew install postgresql
    brew services start postgresql
    ```

2. **Access PostgreSQL CLI by running the following command:**

    ```bash
    psql -U postgres
    ```

3. **Create a database called `your_db_name` and user called `your_db_user`:**
   
    ```bash
    CREATE DATABASE your_db_name;
    CREATE USER your_db_user WITH ENCRYPTED PASSWORD 'your_db_password';
    GRANT ALL PRIVILEGES ON DATABASE your_db_name TO your_db_user;
    ```

4. **Exit PostgreSQL CLI using the following command:**

    ```bash
    \q
    ```

### Option 2: Using pgAdmin 4 (GUI)

1. Open pgAdmin 4 and connect to your PostgreSQL server
2. Right-click on "Databases" → "Create" → "Database..."
3. Enter your database name (e.g., `cms_backend_db`)
4. Go to "Login/Group Roles" → "Create" → "Login/Group Role..."
5. Create a user with login privileges and set a password
6. Grant privileges to the user on your database

### Option 3: Quick Setup Script

For convenience, you can use this one-liner to set up everything:

```bash
# Replace with your preferred values
DB_NAME="cms_backend_db"
DB_USER="cms_user"
DB_PASSWORD="your_secure_password"

createdb $DB_NAME
psql -d $DB_NAME -c "CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';"
psql -d $DB_NAME -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
psql -d $DB_NAME -c "GRANT ALL ON SCHEMA public TO $DB_USER;"
```

### Option 4: Commands for Current .env Configuration

Based on the current `.env` file configuration, run these specific commands:

```bash
createuser cms_user
createdb cms_db

# Set password and grant privileges
psql -d cms_db -c "ALTER USER cms_user WITH PASSWORD 'cms_password';"
psql -d cms_db -c "GRANT ALL PRIVILEGES ON DATABASE cms_db TO cms_user;"
psql -d cms_db -c "GRANT ALL ON SCHEMA public TO cms_user;"
```

**Alternative one-liner for .env configuration:**

```bash
createuser cms_user && createdb cms_db && \
psql -d cms_db -c "ALTER USER cms_user WITH PASSWORD 'cms_password';" && \
psql -d cms_db -c "GRANT ALL PRIVILEGES ON DATABASE cms_db TO cms_user;" && \
psql -d cms_db -c "GRANT ALL ON SCHEMA public TO cms_user;"
```

## Redis Installation and Setup

Redis is required for external caching and multi-instance deployments.

### Option 1: Using Homebrew (macOS)

```bash
brew install redis
brew services start redis

# Verify Redis is running
redis-cli ping
# Should return: PONG
```

### Option 2: Using Package Manager (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install redis-server

# Start Redis service
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Verify Redis is running
redis-cli ping
# Should return: PONG
```

### Option 3: Using Docker

```bash
# Run Redis in Docker container
docker run -d -p 6379:6379 --name cms-redis redis:7-alpine

# Verify Redis is running
docker exec cms-redis redis-cli ping
# Should return: PONG
```

### Option 4: Using Windows (with WSL2)

```bash
sudo apt update
sudo apt install redis-server

# Start Redis manually
sudo service redis-server start

# Or enable auto-start
sudo systemctl enable redis-server

# Verify installation
redis-cli ping
# Should return: PONG
```

## Redis Configuration

### Basic Configuration Notes

- **Default Port**: Redis runs on `localhost:6379` (matches `.env` defaults)
- **Authentication**: No authentication required for local development
- **Persistence**: Redis will automatically persist data to disk
- **Fallback**: The application gracefully falls back to in-memory cache if Redis is unavailable

### Advanced Configuration (Optional)

For production environments, you may want to configure Redis with authentication and custom settings:

1. **Create Redis configuration file:**

    ```bash
    # Create custom Redis config
    sudo nano /etc/redis/redis.conf
    ```

2. **Add security settings:**

    ```conf
    # Require password authentication
    requirepass your_redis_password
    
    # Limit memory usage
    maxmemory 256mb
    maxmemory-policy allkeys-lru
    
    # Enable persistence
    save 900 1
    save 300 10
    save 60 10000
    ```

3. **Restart Redis with custom config:**

    ```bash
    # Stop Redis
    brew services stop redis
    
    # Start with custom config
    redis-server /etc/redis/redis.conf
    ```

4. **Update your .env file:**

    ```bash
    REDIS_HOST=localhost
    REDIS_PORT=6379
    REDIS_PASSWORD=your_redis_password
    REDIS_DB=0
    REDIS_PREFIX=cms_cache:
    ```

## Troubleshooting

### Common PostgreSQL Issues

1. **Connection refused:**
   ```bash
   # Check if PostgreSQL is running
   brew services list | grep postgresql
   
   # Start if not running
   brew services start postgresql
   ```

2. **Permission denied:**
   ```bash
   # Reset PostgreSQL permissions
   sudo -u postgres psql
   ALTER USER postgres PASSWORD 'newpassword';
   ```

3. **Database already exists:**
   ```bash
   dropdb cms_db
   createdb cms_db
   ```

### Common Redis Issues

1. **Redis not starting:**
   ```bash
   # Check Redis status
   brew services list | grep redis
   
   # Check Redis logs
   tail -f /usr/local/var/log/redis.log
   ```

2. **Port conflicts:**
   ```bash
   # Check what's using port 6379
   lsof -i :6379
   
   # Kill conflicting process
   sudo kill -9 <PID>
   ```

3. **Permission issues:**
   ```bash
   # Fix Redis directory permissions
   sudo chown -R $(whoami) /usr/local/var/db/redis/
   ```

## Testing Your Setup

### Verify Database Connection

```bash
# Test database connection
psql -h localhost -U cms_user -d cms_db -c "SELECT version();"
```

### Verify Redis Connection

```bash
# Test Redis connection
redis-cli ping

# Test Redis with authentication (if configured)
redis-cli -a your_redis_password ping

# Check Redis info
redis-cli info server
```

### Test Application Connection

After setting up both PostgreSQL and Redis, test with the CMS application:

```bash
cd starter
go run main.go
```

You should see output similar to:
```
Connected to Redis at localhost:6379 (DB: 0)
Using Redis cache as primary with in-memory fallback
[GIN-debug] Listening and serving HTTP on :8080
```

## Environment Variables

Ensure your `.env` file is properly configured:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=cms_user
DB_PASSWORD=cms_password
DB_NAME=cms_db

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_PREFIX=cms_cache:

ENV=development
PORT=8080
```

## Next Steps

After completing this setup:

1. Return to the main [README.md](README.md) to continue with database migrations
2. Run the application and test the endpoints
3. Use the Redis testing script: `./redis-test.sh`
4. Consider Docker deployment for production: [DOCKER.md](DOCKER.md)

For more information about the project architecture and API usage, see the main [README.md](README.md).