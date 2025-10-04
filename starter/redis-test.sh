#!/bin/bash

# Redis Caching Test Script
# This script demonstrates Redis caching functionality

echo "========================================="
echo "CMS Backend Redis Caching Test"
echo "========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}1. Checking Redis connection...${NC}"
if redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Redis is running${NC}"
    REDIS_AVAILABLE=true
else
    echo -e "${YELLOW}⚠️  Redis not available, will use in-memory cache${NC}"
    REDIS_AVAILABLE=false
fi
echo

echo -e "${BLUE}2. Building CMS Backend with Redis support...${NC}"
if go build -o cms-redis main.go; then
    echo -e "${GREEN}✅ Build successful${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo

echo -e "${BLUE}3. Starting CMS Backend server...${NC}"

echo "Checking for existing processes on port 8080..."
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "Stopping existing server on port 8080..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null
    sleep 2
fi

if [ "$REDIS_AVAILABLE" = true ]; then
    REDIS_HOST=localhost ./cms-redis &
else
    ./cms-redis &
fi
SERVER_PID=$!
echo -e "${GREEN}✅ Server started on port 8080 (PID: $SERVER_PID)${NC}"

sleep 3

make_request() {
    local url=$1
    local description=$2
    
    echo -e "${YELLOW}Making request: $description${NC}"
    
    temp_file=$(mktemp)
    response_file=$(mktemp)
    
    curl -s -D "$temp_file" -o "$response_file" "$url"
    
    cat "$response_file"
    echo
    
    cache_status=$(grep -i "^x-cache:" "$temp_file" | cut -d: -f2 | tr -d ' \r\n' | head -1)
    cache_key=$(grep -i "^x-cache-key:" "$temp_file" | cut -d: -f2 | tr -d ' \r\n' | head -1)
    
    if [ -z "$cache_status" ]; then
        echo "Debug: Raw headers from response:"
        cat "$temp_file"
        echo "Debug: Looking for X-Cache headers:"
        grep -i "x-cache" "$temp_file" || echo "No X-Cache headers found"
    fi
    
    rm -f "$temp_file" "$response_file"
    
    if [ "$cache_status" = "HIT" ]; then
        echo -e "${GREEN}Cache: HIT${NC} (Key: $cache_key)"
    elif [ "$cache_status" = "MISS" ]; then
        echo -e "${YELLOW}Cache: MISS${NC} (Key: $cache_key)"
    elif [ -n "$cache_status" ]; then
        echo -e "${BLUE}Cache: $cache_status${NC} (Key: $cache_key)"
    else
        echo -e "${RED}Cache: No cache headers found${NC}"
    fi
    echo
}

echo -e "${BLUE}4. Demonstrating cache functionality...${NC}"
echo

make_request "http://localhost:8080/api/v1/media" "Get media (first time)"

make_request "http://localhost:8080/api/v1/media" "Get media (second time)"

make_request "http://localhost:8080/api/v1/posts" "Get posts (first time)"

make_request "http://localhost:8080/api/v1/posts" "Get posts (second time)"

echo -e "${BLUE}5. Checking cache statistics...${NC}"
cache_stats=$(curl -s "http://localhost:8080/api/v1/cache/stats")
echo "$cache_stats" | jq '.' 2>/dev/null || echo "$cache_stats"
echo

echo -e "${BLUE}6. Testing cache invalidation...${NC}"
echo "Invalidating media cache..."
curl -s -X POST "http://localhost:8080/api/v1/cache/invalidate/media" | jq '.' 2>/dev/null
echo

make_request "http://localhost:8080/api/v1/media" "Get media (after invalidation)"

echo -e "${BLUE}7. Checking cache health...${NC}"
cache_health=$(curl -s "http://localhost:8080/api/v1/cache/health")
echo "$cache_health" | jq '.' 2>/dev/null || echo "$cache_health"
echo

echo -e "${BLUE}8. Running performance test...${NC}"
echo "Making 10 requests to test cache performance..."

start_time=$(date +%s%N)
for i in {1..10}; do
    curl -s "http://localhost:8080/api/v1/media" > /dev/null
done
end_time=$(date +%s%N)

duration=$(( (end_time - start_time) / 1000000 ))
echo -e "${GREEN}✅ 10 requests completed in ${duration}ms${NC}"
echo "Average: $((duration / 10))ms per request"
echo

echo -e "${BLUE}9. Final cache statistics...${NC}"
final_stats=$(curl -s "http://localhost:8080/api/v1/cache/stats")
echo "$final_stats" | jq '.' 2>/dev/null || echo "$final_stats"
echo
echo "2. Running comprehensive test suite..."
go test ./... -cover
if [ $? -eq 0 ]; then
    echo "✅ All tests pass with performance optimizations"
else
    echo "❌ Tests failed"
    exit 1
fi

echo -e "${BLUE}10. Cleanup...${NC}"
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
rm -f cms-redis

echo -e "${GREEN}✅ Test completed successfully!${NC}"
echo
echo "========================================="
echo "Redis Caching Test Summary"
echo "========================================="
echo "Features demonstrated:"
echo "• Cache HIT/MISS behavior"
echo "• Cache key generation"
echo "• Cache invalidation"
echo "• Performance monitoring"
echo "• Health checks"
echo "• Statistics tracking"
echo
if [ "$REDIS_AVAILABLE" = true ]; then
    echo -e "${GREEN}✅ Redis cache was used${NC}"
else
    echo -e "${YELLOW}⚠️  In-memory cache was used (Redis not available)${NC}"
fi
echo "========================================="