package integration

import (
	"cms-backend/models"
	"cms-backend/routes"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB *gorm.DB
	router *gin.Engine
)

// TODO: Import required packages for:
// - Database (gorm, postgres driver)
// - Gin framework
// - Testing
// - Logging
// - OS operations
// - Your application packages (models, routes)

// TODO: Define package-level variables for:
// - Test database connection
// - Gin router instance

/*
INTEGRATION TEST SETUP GUIDE

This file sets up the integration test environment for your CMS backend.
It handles database connections, schema migrations, and cleanup.

Key Components:
1. Test database connection
2. Router setup
3. Schema migrations
4. Test cleanup
*/

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func setup() {
	//TODO: Implement setup
	// STEP 1: Configure Gin
	// - Set Gin to test mode for integration testing
	gin.SetMode(gin.TestMode)

	// STEP 2: Database Connection
	// - Define test database connection string
	// - Connect to test database using GORM
	// - Store connection in testDB variable
	// - Handle connection errors
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=cms_test port=5432 sslmode=disable"
	}
	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// STEP 3: Schema Migration
	// - Migrate all model schemas:
	//   * Media
	//   * Page
	//   * Post
	//   * Any join tables
	if err := testDB.AutoMigrate(&models.Media{}, &models.Page{}, &models.Post{}); err != nil {
		log.Fatalf("Failed to migrate schemas: %v", err)
	}
	// Migrate join table for Post-Media
	if err := testDB.SetupJoinTable(&models.Post{}, "Media", &models.PostMedia{}); err != nil {
		log.Fatalf("Failed to setup join table: %v", err)
	}

	// STEP 4: Router Setup
	// - Initialize Gin router
	// - Set up routes with test database
	router = gin.New()
	routes.InitializeRoutes(router, testDB)
}

func cleanup() {
	//TODO: Implement cleanup
	// STEP 1: Database Cleanup
	// - Get underlying SQL database
	// - Drop all tables in correct order:
	//   1. Junction tables (post_media)
	//   2. Main tables (posts, media, pages)
	sqlDB, err := testDB.DB()
	if err != nil {
		log.Printf("Failed to get sql.DB: %v", err)
		return
	}

	if err := testDB.Migrator().DropTable("post_media"); err != nil {
		log.Printf("Failed to drop post_media: %v", err)
	}

	for _, tbl := range []string{"posts", "media", "pages"} {
		if err := testDB.Migrator().DropTable(tbl); err != nil {
			log.Printf("Failed to drop table %s: %v", tbl, err)
		}
	}

	// STEP 2: Connection Cleanup
	// - Close database connection
	// - Handle any cleanup errors
	sqlDB.Close()
}

func clearTables() {
	//TODO: Implement clearTables
	// STEP 1: Data Cleanup
	// - Delete all data from tables in correct order:
	//   1. Junction tables first
	//   2. Main tables next
	// - Maintain referential integrity
	// Delete data in correct order
	testDB.Exec("DELETE FROM post_media")
	testDB.Exec("DELETE FROM posts")
	testDB.Exec("DELETE FROM media")
	testDB.Exec("DELETE FROM pages")
}

/*
TESTING HINTS:
1. Database Connection:
   - Use a separate test database
   - Consider environment variables for credentials
   - Handle connection errors properly

2. Table Management:
   - Drop tables in correct order (foreign key constraints)
   - Clear data between tests
   - Consider using transactions for tests

3. Error Handling:
   - Log setup/cleanup errors
   - Ensure proper resource cleanup
   - Handle database operation errors

4. Best Practices:
   - Use constants for connection strings
   - Consider test helper functions
   - Add proper logging for debugging
   - Document any required setup steps
*/
