#!/bin/bash

# Migration script for CMS Backend Project
# Usage: ./migrate.sh [up|down|version|create]

if [ -f .env ]; then
    export $(cat .env | grep -v ^# | xargs)
fi

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-cms_backend}

DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

case $1 in
    "up")
        echo "🚀 Running migrations up..."
        migrate -database "$DATABASE_URL" -path ./migrations up
        ;;
    "down")
        echo "⬇️  Rolling back migrations..."
        if [ -n "$2" ]; then
            migrate -database "$DATABASE_URL" -path ./migrations down $2
        else
            migrate -database "$DATABASE_URL" -path ./migrations down 1
        fi
        ;;
    "version")
        echo "📋 Checking migration version..."
        migrate -database "$DATABASE_URL" -path ./migrations version
        ;;
    "force")
        if [ -n "$2" ]; then
            echo "⚠️  Forcing migration to version $2..."
            migrate -database "$DATABASE_URL" -path ./migrations force $2
        else
            echo "❌ Please provide version number: ./migrate.sh force <version>"
        fi
        ;;
    "create")
        if [ -n "$2" ]; then
            echo "📝 Creating new migration: $2"
            migrate create -ext sql -dir ./migrations -seq $2
        else
            echo "❌ Please provide migration name: ./migrate.sh create <migration_name>"
        fi
        ;;
    "drop")
        echo "⚠️  This will DROP ALL TABLES! Are you sure? (y/N)"
        read confirmation
        if [ "$confirmation" = "y" ] || [ "$confirmation" = "Y" ]; then
            migrate -database "$DATABASE_URL" -path ./migrations drop -f
        else
            echo "❌ Operation cancelled"
        fi
        ;;
    *)
        echo "📖 CMS Backend Migration Helper"
        echo ""
        echo "Usage: ./migrate.sh [command] [options]"
        echo ""
        echo "Commands:"
        echo "  up                 - Apply all pending migrations"
        echo "  down [N]          - Rollback N migrations (default: 1)"
        echo "  version           - Show current migration version"
        echo "  force <version>   - Force set version without running migration"
        echo "  create <name>     - Create new migration files"
        echo "  drop              - Drop all tables (WARNING: destructive)"
        echo ""
        echo "Examples:"
        echo "  ./migrate.sh up"
        echo "  ./migrate.sh down"
        echo "  ./migrate.sh down 2"
        echo "  ./migrate.sh version"
        echo "  ./migrate.sh create add_user_table"
        echo ""
        echo "Environment variables (from .env):"
        echo "  DB_HOST=$DB_HOST"
        echo "  DB_PORT=$DB_PORT"
        echo "  DB_USER=$DB_USER"
        echo "  DB_NAME=$DB_NAME"
        ;;
esac