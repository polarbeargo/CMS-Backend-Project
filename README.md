# CMS Backend Project
## Project Objectives

The objective of this project is to build a fully functional backend API with CRUD capabilities for managing pages, posts, and media content in a content management system (CMS). By completing this project, you will:

- Apply Go programming skills to build a RESTful API.
- Utilize the Gin web framework for efficient routing and middleware management.
- Implement database interactions using GORM with PostgreSQL.
- Manage database schema changes using migrations.
- Write unit and integration tests to ensure code quality and reliability.
- Configure environment variables for different deployment environments.

The final deliverable is a backend API with comprehensive test coverage that follows best practices in backend development and demonstrates your ability to:

- Build scalable backend services using Go, one of the most in-demand programming languages in the industry.
- Work with relational databases and ORMs, a critical skill for backend developers.
- Implement RESTful APIs, which are foundational to modern web development.
- Write robust tests, showcasing your commitment to code quality and reliability.
- Manage environment configurations and understand the importance of secure credential handling.

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go 1.16 or higher](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/) (Optional - for external caching)
- [Git](https://git-scm.com/)
- [golang-migrate](https://github.com/golang-migrate)
- [pgAdmin 4](https://www.pgadmin.org/download/) (Optional)
- [Docker & Docker Compose](https://www.docker.com/) (Optional - for containerized deployment)

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

    Create a `.env` file. In the project root directory, create a `.env` file to store your development environment variables:

    ```bash
    cp .env.example .env
    ```

    Open the `.env` file and replace the placeholder values with your actual database credentials:

    ```bash
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=your_db_user
    DB_PASSWORD=your_db_password
    DB_NAME=your_db_name
    ENV=development
    ```

    Note: Ensure that the .env file is included in your .gitignore to prevent sensitive information from being committed to version control. Add `.env` to `.gitignore`.

4. **Create Databases**

    You can set up your PostgreSQL database using either the PostgreSQL CLI or pgAdmin 4.

    **Option 1: Using PostgreSQL CLI**

    1. Start PostgreSQL service using the following commands:

        ```bash
        brew install postgresql
        brew services start postgresql
        ```

    2. Access PostgreSQL CLI by running the following command:

        ```bash
        psql -U postgres
        ```

    3. Create a database called `your_db_name` and user called `your_db_user`.
       
        ```bash
        CREATE DATABASE your_db_name;
        CREATE USER your_db_user WITH ENCRYPTED PASSWORD 'your_db_password';
        GRANT ALL PRIVILEGES ON DATABASE your_db_name TO your_db_user;
        ```

    4. Exit PostgreSQL CLI using the following command:

        ```bash
        \q
        ```

    **Option 2: Using pgAdmin 4 (GUI)**

    1. Open pgAdmin 4 and connect to your PostgreSQL server
    2. Right-click on "Databases" → "Create" → "Database..."
    3. Enter your database name (e.g., `cms_backend_db`)
    4. Go to "Login/Group Roles" → "Create" → "Login/Group Role..."
    5. Create a user with login privileges and set a password
    6. Grant privileges to the user on your database

    **Option 3: Quick Setup Script**

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

## System Architecture Overview

### Architecture Diagram

![CMS Backend Architecture Diagram](starter/docs/CMS%20Backend.png)

### API Request Flow

![API Request Flow Diagram](starter/docs/api_request_flow.png)

### Security & Performance Features

![Security Features Diagram](starter/docs/security.png)

### Domain Model Relationships

![Domain Models Diagram](starter/docs/domain_models.png)

### Database Schema

![Database Schema Diagram](starter/docs/db_schema.png)

### Package Dependency Flow

```text
 Dependency Flow: main → routes → controllers → models               
                        ↓                                            
                    utils (db, response)                            
```

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

