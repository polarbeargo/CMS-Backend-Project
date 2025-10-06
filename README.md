# CMS Backend Project
A high-performance Content Management System (CMS) backend built with Go, Gin, GORM, PostgreSQL, and Redis caching.
## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go 1.16 or higher](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/) (For external caching)
- [Git](https://git-scm.com/)
- [golang-migrate](https://github.com/golang-migrate)
- [pgAdmin 4](https://www.pgadmin.org/download/) (Optional)
- [Docker & Docker Compose](https://www.docker.com/) (For containerized deployment)

### Installation steps

1. **Clone the repository**

    ```bash
    git clone https://github.com/polarbeargo/CMS-Backend-Project.git
    cd starter
    ```
2. **Install Go dependencies** 

    ```bash
    go mod download
    ```

3. **Set Up Environment Variables**

    Create a `.env` file. In the project root directory to store your development environment variables:

    ```bash
    cp .env.example .env
    ```

    Note: Ensure that the .env file is included in your .gitignore to prevent sensitive information from being committed to version control. Add `.env` to `.gitignore`.

4. **Setup Database and Redis**

    For detailed database and Redis setup instructions, see [Database and Redis Setup Guide](SETUP.md).

5. **Run Database Migrations**

   After setting up the database, run the migrations to create the required tables:

   ```bash
   ./migrate.sh up
   ```

   Or manually:

   ```bash
   migrate -path migrations -database "postgres://your_db_user:your_db_password@localhost:5432/your_db_name?sslmode=disable" up
   ```

6. **Run the Application:**

   ```bash
   go run main.go
   ```

7. **Test Endpoints Using cURL or Postman:**

   ```bash
   curl -X GET http://localhost:8080/api/v1/posts 
   ```

   - Create a New Post:

   ```bash
   curl -X POST http://localhost:8080/api/v1/posts \
     -H "Content-Type: application/json" \
     -d '{"title": "First Post", "content": "This is the content of the first post.", "author": "Admin"}'
   ```

8. **Run Tests:**

   ```bash
   go test ./...
   ```

   - Test coverage report:

   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

9. **Test Redis Caching System:**

    The project includes a comprehensive Redis caching test script to demonstrate and validate the caching functionality:

    ```bash
    cd starter
    ./redis-test.sh
    ```

## Docker Deployment

For containerized deployment using Docker and Docker Compose, see [Docker Deployment Guide](DOCKER.md).

The Docker setup includes:
- **Multi-service architecture** with PostgreSQL, Redis, and the CMS backend
- **Development and production profiles** with different configurations
- **Health checks and monitoring** for all services
- **Data persistence** with automatic volume management
- **Redis Commander GUI** for development
- **Nginx load balancer** for production scaling

**Quick Start:**
```bash
cd starter
docker-compose up -d
```

For detailed instructions, troubleshooting, and advanced configurations, see [DOCKER.md](DOCKER.md).

## System Architecture Overview

### Architecture Diagram

This diagram provides a high-level view of the entire CMS backend system architecture.

```mermaid
graph TB
    subgraph "Client Layer"
        Web[Web Applications]
        Mobile[Mobile Apps]
        API[API Clients]
    end
    
    subgraph "Load Balancer"
        LB[Load Balancer<br/>nginx/HAProxy]
    end
    
    subgraph "Application Layer"
        direction TB
        subgraph "Instance 1"
            App1[CMS Backend<br/>Go + Gin]
        end
        subgraph "Instance 2" 
            App2[CMS Backend<br/>Go + Gin]
        end
        subgraph "Instance N"
            AppN[CMS Backend<br/>Go + Gin]
        end
    end
    
    subgraph "Caching Layer"
        Redis[(Redis Cache<br/>Shared across instances)]
        subgraph "Local Caches"
            Mem1[In-Memory<br/>Instance 1]
            Mem2[In-Memory<br/>Instance 2]
            MemN[In-Memory<br/>Instance N]
        end
    end
    
    subgraph "Data Layer"
        DB[(PostgreSQL<br/>Primary Database)]
        Migrations[Database Migrations<br/>golang-migrate]
    end
    
    subgraph "Infrastructure"
        Docker[Docker Containers]
        Compose[Docker Compose<br/>Multi-service orchestration]
    end
    
    Web --> LB
    Mobile --> LB
    API --> LB
    
    LB --> App1
    LB --> App2
    LB --> AppN
    
    App1 <--> Redis
    App2 <--> Redis
    AppN <--> Redis
    
    App1 -.-> Mem1
    App2 -.-> Mem2
    AppN -.-> MemN
    
    App1 <--> DB
    App2 <--> DB
    AppN <--> DB
    
    Migrations --> DB
    
    App1 -.-> Docker
    App2 -.-> Docker
    AppN -.-> Docker
    Redis -.-> Docker
    DB -.-> Docker
    
    Docker --> Compose
```

## Component Responsibilities

### Application Layer
- **Gin Framework**: HTTP routing and middleware
- **GORM**: Database ORM and migrations
- **Redis Cache Middleware**: Response caching for GET requests
- **Controllers**: Business logic for Pages, Posts, Media, and Cache management

### Caching Strategy  
- **Redis (Primary)**: Shared cache across all application instances
- **In-Memory (Fallback)**: Local cache per instance when Redis unavailable
- **Automatic Invalidation**: Cache cleared on data modifications
- **TTL Management**: 5-minute default expiration for cached responses

### Data Persistence
- **PostgreSQL**: Primary data store with ACID compliance
- **Migrations**: Version-controlled schema changes
- **Connection Pooling**: Efficient database connection management

### Deployment Architecture
- **Multi-Instance**: Horizontal scaling with multiple app instances
- **Container-Based**: Docker containers for consistent environments  
- **Service Orchestration**: Docker Compose for local development
- **Load Balancing**: Distributes traffic across instances

## Scalability Features

1. **Horizontal Scaling**: Add more application instances behind load balancer
2. **Shared Cache**: Redis cache shared across all instances for consistency
3. **Database Connection Pooling**: Efficient resource utilization
4. **Stateless Design**: Each instance can handle any request
5. **Cache Fallback**: System remains functional even if Redis fails

## Development Workflow

1. **Database Setup**: Run migrations to create schema
2. **Redis Setup**: Start Redis server for caching
3. **Application Start**: Launch one or more app instances
4. **Monitoring**: Check cache statistics and system health

# Caching Architecture

This diagram shows the caching subsystem with Redis primary cache and in-memory fallback.

```mermaid
classDiagram
    direction TB
    
    class CacheInterface {
        <<interface>>
        +Get(key string) (*CacheItem, bool)
        +Set(key string, data []byte, headers map[string]string, ttl time.Duration) error
        +Delete(key string) error
        +Clear() error
        +InvalidatePattern(pattern string) error
        +GetStats() CacheStats
        +ResetStats() error
    }
    
    class CacheItem {
        +[]byte Data
        +map[string]string Headers
        +time.Time ExpiresAt
    }
    
    class CacheStats {
        +int64 Hits
        +int64 Misses
        +float64 HitRatio
        +int64 KeyCount
        +string CacheType
        +time.Time LastCleanup
    }
    
    class RedisCache {
        -*redis.Client client
        -string prefix
        -CacheStats stats
        +Get(key string) (*CacheItem, bool)
        +Set(key string, data []byte, headers map[string]string, ttl time.Duration) error
        +Delete(key string) error
        +Clear() error
        +InvalidatePattern(pattern string) error
        +GetStats() CacheStats
        +ResetStats() error
        +NewRedisCache() (*RedisCache, error)
    }
    
    class InMemoryCache {
        -map[string]*CacheItem items
        -sync.RWMutex mutex
        +Get(key string) (*CacheItem, bool)
        +Set(key string, data []byte, headers map[string]string, ttl time.Duration) error
        +Delete(key string) error
        +Clear() error
        +InvalidatePattern(pattern string) error
        +GetStats() CacheStats
        +ResetStats() error
        +NewInMemoryCache() *InMemoryCache
        -cleanup()
    }
    
    class CacheManager {
        -CacheInterface primary
        -CacheInterface fallback
        -bool useFallback
        +Get(key string) (*CacheItem, bool)
        +Set(key string, data []byte, headers map[string]string, ttl time.Duration) error
        +Delete(key string) error
        +Clear() error
        +InvalidatePattern(pattern string) error
        +GetStats() CacheStats
        +ResetStats() error
        +NewCacheManager() *CacheManager
    }
    
    class responseWriter {
        +gin.ResponseWriter
        +[]byte body
        +map[string]string headers
        +Write(data []byte) (int, error)
    }
    
    CacheInterface <|.. RedisCache : implements
    CacheInterface <|.. InMemoryCache : implements
    CacheManager o--> CacheInterface : primary
    CacheManager o--> CacheInterface : fallback
    RedisCache ..> CacheItem : uses
    InMemoryCache ..> CacheItem : uses
    CacheManager ..> CacheStats : returns
    responseWriter ..> gin.ResponseWriter : embeds
```

## Cache Strategy

1. **Dual Cache System**: Redis as primary, in-memory as fallback
2. **Automatic Fallback**: If Redis fails, falls back to in-memory cache
3. **TTL Management**: Time-based expiration for all cache entries
4. **Pattern Invalidation**: Can invalidate cache entries by pattern matching
5. **Statistics Tracking**: Monitors hits, misses, and hit ratios

## Cache Invalidation Functions

- `InvalidateMediaCache()` - Clears media-related cache entries
- `InvalidatePostCache()` - Clears post-related cache entries  
- `InvalidatePageCache()` - Clears page-related cache entries

### API Routing Architecture

This diagram shows the HTTP routing structure and controller organization.

```mermaid
graph TB
    subgraph "HTTP Layer"
        Client[Client Applications]
        Router[Gin Router Engine]
        MW[Redis Cache Middleware]
    end
    
    subgraph "API Routes (/api/v1)"
        direction TB
        PageRoutes["Pages
        GET /pages
        GET /pages/:id
        POST /pages
        PUT /pages/:id
        DELETE /pages/:id"]
        PostRoutes["Posts
        GET /posts
        GET /posts/:id
        POST /posts
        PUT /posts/:id
        DELETE /posts/:id"]
        MediaRoutes["Media
        GET /media
        GET /media/:id
        POST /media
        DELETE /media/:id"]
        CacheRoutes["Cache Management
        GET /cache/stats
        POST /cache/clear
        POST /cache/invalidate
        POST /cache/invalidate/:resource
        POST /cache/warmup
        GET /cache/health"]
    end
    
    subgraph "Controllers"
        PageController["Page Controller
        - GetPages()
        - GetPage()
        - CreatePage()
        - UpdatePage()
        - DeletePage()"]
        PostController["Post Controller
        - GetPosts()
        - GetPost()
        - CreatePost()
        - UpdatePost()
        - DeletePost()"]
        MediaController["Media Controller
        - GetMedia()
        - GetMediaByID()
        - CreateMedia()
        - DeleteMedia()"]
        CacheController["Cache Controller
        - GetCacheStats()
        - ClearCache()
        - InvalidateCache()
        - InvalidateResourceCache()
        - WarmupCache()
        - CacheHealth()"]
    end
    
    subgraph "Data Layer"
        DB["Database
        GORM"]
        Cache["Redis Cache
        + In-Memory Fallback"]
    end
    
    Client --> Router
    Router --> MW
    MW --> PageRoutes
    MW --> PostRoutes  
    MW --> MediaRoutes
    MW --> CacheRoutes
    
    PageRoutes --> PageController
    PostRoutes --> PostController
    MediaRoutes --> MediaController
    CacheRoutes --> CacheController
    
    PageController <--> DB
    PostController <--> DB
    MediaController <--> DB
    
    MW <--> Cache
    PageController -.-> Cache
    PostController -.-> Cache
    MediaController -.-> Cache
```
## Request Flow

1. **Client Request** → Gin Router
2. **Middleware Processing** → Redis Cache Middleware (for GET requests)
3. **Route Matching** → Appropriate Controller
4. **Business Logic** → Controller methods
5. **Data Access** → Database via GORM
6. **Cache Management** → Automatic invalidation on write operations

## Middleware Features

- **Caching**: Only caches GET requests with 2xx status codes
- **Cache Headers**: Adds X-Cache (HIT/MISS) and X-Cache-Key headers
- **Selective Caching**: Skips /admin and /auth endpoints
- **Response Capture**: Uses custom responseWriter to capture response data 

## Request Flow Sequences

This document shows the sequence diagrams for key operations in the CMS system.

## GET Request with Caching

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant R as Gin Router
    participant MW as Cache Middleware
    participant CM as Cache Manager
    participant RC as Redis Cache
    participant PC as Controller
    participant DB as Database

    C->>R: GET /api/v1/posts
    R->>MW: Process Request
    MW->>CM: Get(cacheKey)
    CM->>RC: Get(key)
    
    alt Cache HIT
        RC-->>CM: CacheItem found
        CM-->>MW: Return cached data
        MW-->>C: 200 OK (X-Cache: HIT)
    else Cache MISS
        RC-->>CM: Key not found
        CM-->>MW: Cache miss
        MW->>PC: Forward to GetPosts()
        PC->>DB: Query posts with media
        DB-->>PC: Return result set
        PC-->>MW: 200 OK + JSON response
        MW->>CM: Set(cacheKey, response, headers, TTL)
        CM->>RC: Store in Redis
        RC-->>CM: Stored successfully
        MW-->>C: 200 OK (X-Cache: MISS)
    end
```

## POST Request with Cache Invalidation

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant R as Gin Router
    participant PC as Post Controller
    participant DB as Database
    participant INV as InvalidatePostCache
    participant CM as Cache Manager
    participant RC as Redis Cache

    C->>R: POST /api/v1/posts
    Note over R: No caching for POST requests
    R->>PC: CreatePost(requestBody)
    PC->>PC: Validate input
    PC->>DB: Begin transaction
    PC->>DB: INSERT INTO posts
    PC->>DB: INSERT INTO post_media (if media)
    PC->>DB: Commit transaction
    DB-->>PC: Post created successfully
    
    PC->>INV: InvalidatePostCache()
    INV->>CM: InvalidatePattern("posts")
    CM->>RC: Delete keys matching "*posts*"
    RC-->>CM: Keys deleted
    CM-->>INV: Invalidation complete
    
    PC-->>C: 201 Created + new post data
```

## Cache Management Operations

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant CC as Cache Controller
    participant CM as Cache Manager
    participant RC as Redis Cache
    participant IM as In-Memory Cache

    alt Get Cache Statistics
        C->>CC: GET /cache/stats
        CC->>CM: GetStats()
        CM->>RC: GetStats()
        RC-->>CM: CacheStats object
        CM-->>CC: Statistics data
        CC-->>C: 200 OK + stats JSON
    
    else Clear All Cache
        C->>CC: POST /cache/clear
        CC->>CM: Clear()
        CM->>RC: Clear Redis cache
        CM->>IM: Clear in-memory cache
        RC-->>CM: Cleared
        IM-->>CM: Cleared
        CM-->>CC: Success
        CC-->>C: 200 OK + success message
    
    else Invalidate by Pattern
        C->>CC: POST /cache/invalidate?pattern=posts
        CC->>CM: InvalidatePattern("posts")
        CM->>RC: Delete matching keys
        RC-->>CM: Keys deleted
        CM-->>CC: Success
        CC-->>C: 200 OK + invalidation result
    end
```

## Multi-Instance Cache Consistency

```mermaid
sequenceDiagram
    autonumber
    participant I1 as Instance 1
    participant I2 as Instance 2
    participant I3 as Instance 3
    participant Redis as Shared Redis Cache
    participant DB as Database

    Note over I1,I3: All instances share the same Redis cache
    
    I1->>Redis: GET posts (cache miss)
    Redis-->>I1: No data found
    I1->>DB: Query posts from database
    DB-->>I1: Return posts data
    I1->>Redis: Cache posts data (TTL: 5min)
    
    Note over I2: User requests posts from different instance
    I2->>Redis: GET posts (cache hit)
    Redis-->>I2: Return cached posts data
    
    Note over I3: User creates new post via third instance
    I3->>DB: INSERT new post
    I3->>Redis: InvalidatePattern("posts")
    Redis-->>I3: Cache cleared
    
    Note over I1,I2: Next request to any instance gets fresh data
    I1->>Redis: GET posts (cache miss after invalidation)
    Redis-->>I1: No data found
    I1->>DB: Query fresh posts data
    DB-->>I1: Return updated posts (including new post)
    I1->>Redis: Cache updated posts data
```

## Key Benefits

1. **Performance**: Cached responses serve instantly without database queries
2. **Consistency**: Cache invalidation ensures all instances see fresh data
3. **Reliability**: Automatic fallback to in-memory cache if Redis fails
4. **Monitoring**: Built-in statistics and health endpoints for observability

## Domain Model (Entity Relationships)

This diagram shows the core data entities and their relationships in the CMS system.

```mermaid
erDiagram
    Page {
        uint ID PK
        string Title
        string Content
        time CreatedAt
        time UpdatedAt
    }
    
    Media {
        uint ID PK
        string URL
        string Type
        time CreatedAt
        time UpdatedAt
    }
    
    Post {
        uint ID PK
        string Title
        string Content
        string Author
        time CreatedAt
        time UpdatedAt
    }
    
    PostMedia {
        uint PostID PK,FK
        uint MediaID PK,FK
    }
    
    Post ||--o{ PostMedia : "has"
    Media ||--o{ PostMedia : "referenced_by"
    Post }o--o{ Media : "many2many via PostMedia"
```

## Key Relationships

- **Pages**: Independent content entities (no relationships)
- **Posts**: Can have multiple media attachments via many-to-many relationship
- **Media**: Can be referenced by multiple posts
- **PostMedia**: Junction table implementing the many-to-many relationship

## Business Rules

1. Pages are standalone content (like static pages)
2. Posts can have zero or more media attachments
3. Media files can be reused across multiple posts
4. All entities have automatic timestamps (CreatedAt, UpdatedAt)

## Security & Performance Features

This diagram shows the comprehensive security and performance features implemented in the CMS system.

```mermaid
graph TB
    subgraph "Security Layer"
        direction TB
        subgraph "Input Validation"
            Validator[Go Validator
            - Required fields
            - Data type validation
            - Format validation]
            Sanitization[Input Sanitization
            - SQL injection prevention
            - XSS protection
            - Data cleaning]
        end
        
        subgraph "Database Security"
            GORM[GORM ORM
            - Prepared statements
            - SQL injection protection
            - Type safety]
            Transactions[Database Transactions
            - ACID compliance
            - Rollback on errors
            - Data consistency]
        end
        
        subgraph "Environment Security"
            EnvVars[Environment Variables
            - Secure credential storage
            - .env file exclusion
            - Runtime configuration]
            Secrets[Secret Management
            - Database passwords
            - Redis credentials
            - API keys]
        end
    end
    
    subgraph "Performance Layer"
        direction TB
        subgraph "Caching Strategy"
            RedisCache[Redis Cache
            - Primary cache layer
            - Shared across instances
            - TTL management]
            MemoryCache[In-Memory Cache
            - Fallback mechanism
            - Local per instance
            - Automatic cleanup]
            CacheInvalidation[Smart Invalidation
            - Pattern-based clearing
            - Automatic on writes
            - Resource-specific]
        end
        
        subgraph "Database Optimization"
            ConnPool[Connection Pooling
            - Efficient resource usage
            - Concurrent connections
            - Timeout management]
            Indexing[Database Indexing
            - Primary key optimization
            - Foreign key indexing
            - Query performance]
            Pagination[Smart Pagination
            - Limit result sets
            - Offset optimization
            - Memory efficiency]
        end
        
        subgraph "API Performance"
            Middleware[Efficient Middleware
            - Minimal overhead
            - Selective caching
            - Fast routing]
            Compression[Response Optimization
            - JSON serialization
            - Header management
            - Content-Type handling]
            Monitoring[Performance Monitoring
            - Cache hit ratios
            - Response times
            - Health checks]
        end
    end
    
    subgraph "Scalability Features"
        direction TB
        MultiInstance[Multi-Instance Support
        - Horizontal scaling
        - Load balancing ready
        - Stateless design]
        SharedCache[Shared Cache Layer
        - Cross-instance consistency
        - Synchronized invalidation
        - High availability]
        ContainerReady[Container Support
        - Docker deployment
        - Environment isolation
        - Orchestration ready]
    end
    
    subgraph "Reliability Features"
        direction TB
        ErrorHandling[Error Handling
        - Graceful degradation
        - Structured error responses
        - Logging integration]
        Fallbacks[Fallback Mechanisms
        - Cache fallback
        - Database retry logic
        - Service resilience]
        HealthChecks[Health Monitoring
        - Cache health endpoints
        - Database connectivity
        - System status]
    end
    
    Validator --> GORM
    Sanitization --> GORM
    GORM --> Transactions
    EnvVars --> Secrets
    
    RedisCache -.-> MemoryCache
    CacheInvalidation --> RedisCache
    ConnPool --> Indexing
    Indexing --> Pagination
    
    Middleware --> Compression
    Compression --> Monitoring
    
    MultiInstance --> SharedCache
    SharedCache --> ContainerReady
    
    ErrorHandling --> Fallbacks
    Fallbacks --> HealthChecks
    
    RedisCache --> SharedCache
    ConnPool --> MultiInstance
```

#### Security Features

**Input Validation & Sanitization:**
- Go Validator with comprehensive struct tag validation
- Required field validation and type safety
- SQL injection prevention through GORM prepared statements
- XSS protection with input sanitization

**Database Security:**
- ACID-compliant transactions with automatic rollback
- Prepared statements for all database queries
- Secure credential management via environment variables
- Minimal database user privileges

**Environment Security:**
- Secret management with .env files (excluded from git)
- Runtime configuration isolation
- Secure storage of database passwords and API keys

#### Performance Features

**Multi-Layer Caching:**
- Redis primary cache shared across instances
- In-memory fallback cache for high availability
- Smart invalidation with pattern-based clearing
- TTL management (5-minute default expiration)

**Database Optimization:**
- Connection pooling for efficient resource usage
- Indexed queries with proper foreign key relationships
- Smart pagination to prevent memory issues
- Optimized eager loading with GORM preloading

**API Performance:**
- Efficient Gin framework routing with minimal overhead
- Automatic response caching for GET requests
- JSON optimization with proper content types
- Cache headers for client-side optimization

#### Scalability & Reliability

**Horizontal Scaling:**
- Stateless application design ready for load balancing
- Shared Redis cache for cross-instance consistency
- Container support with Docker deployment
- Multi-instance coordination with synchronized cache invalidation

**Reliability Features:**
- Graceful degradation with cache fallback mechanisms
- Structured error handling with proper HTTP status codes
- Health monitoring endpoints for system status
- Built-in performance metrics and observability

### RESTful API Endpoints

**Base URL:** `http://localhost:8080/api/v1`

| Resource | Method | Path       | Description                                   |
|----------|--------|------------|-----------------------------------------------|
| **Pages** | GET    | /pages     | List all pages (paginated, filterable)        |
|          | GET    | /pages/1   | Get specific page by ID                       |
|          | POST   | /pages     | Create new page                               |
|          | PUT    | /pages/1   | Update existing page                          |
|          | DELETE | /pages/1   | Delete page by ID                             |
| **Posts** | GET    | /posts     | List all posts (with media, paginated)        |
|          | GET    | /posts/1   | Get specific post by ID (with media)          |
|          | POST   | /posts     | Create new post (with media association)       |
|          | PUT    | /posts/1   | Update existing post                          |
|          | DELETE | /posts/1   | Delete post by ID                             |
| **Media** | GET    | /media     | List all media files (paginated)              |
|          | GET    | /media/1   | Get specific media by ID                      |
|          | POST   | /media     | Upload/create new media entry                 |
|          | DELETE | /media/1   | Delete media by ID                            |
| **Cache** | GET    | /cache/stats | Get cache statistics and performance metrics |
|          | GET    | /cache/health | Check cache system health                   |
|          | POST   | /cache/clear | Clear all cache entries                     |
|          | POST   | /cache/invalidate | Invalidate cache by pattern           |
|          | POST   | /cache/invalidate/media | Invalidate media cache          |
|          | POST   | /cache/invalidate/posts | Invalidate posts cache          |
|          | POST   | /cache/invalidate/pages | Invalidate pages cache          |


**Query Parameters (Available on GET endpoints):**

- `page=1`: Pagination (specifies the page number)
- `page_size=10`: Number of items per page (maximum 100)
- `sort_by=created_at`: Sort field (`title`, `created_at`, `updated_at`)
- `sort_order=desc`: Sort order (`asc`, `desc`)
- `search=keyword`: Search by keyword in title and content
- `title=filter`: Filter by title (for posts/pages)
- `author=filter`: Filter by author (for posts)