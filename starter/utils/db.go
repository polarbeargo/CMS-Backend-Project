package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBResources struct {
	GormDB  *gorm.DB
	PgxPool *pgxpool.Pool
}

func ConnectDBWithRetry(maxRetries int, retryDelay time.Duration) (*DBResources, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Database connection attempt %d/%d", attempt, maxRetries)

		dbRes, err := ConnectDB()
		if err == nil {
			log.Println("Database connection successful")
			return dbRes, nil
		}

		lastErr = err
		log.Printf("Database connection failed (attempt %d/%d): %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, lastErr)
}

func ConnectDB() (*DBResources, error) {
	env := os.Getenv("APP_ENV")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	sslMode := "disable"
	if env == "production" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort, sslMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 60 * time.Minute
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pgxPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pgxpool: %w", err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		pgxPool.Close()
		return nil, fmt.Errorf("pgxpool ping failed: %w", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		pgxPool.Close()
		return nil, fmt.Errorf("failed to connect to database (gorm): %w", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		pgxPool.Close()
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		pgxPool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	return &DBResources{
		GormDB:  gormDB,
		PgxPool: pgxPool,
	}, nil
}

func CloseDatabase(dbRes *DBResources) {
	if dbRes == nil {
		return
	}
	if dbRes.PgxPool != nil {
		dbRes.PgxPool.Close()
	}
	if dbRes.GormDB != nil {
		if sqlDB, err := dbRes.GormDB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}
