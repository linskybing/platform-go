#!/bin/bash
# Validate Redis configuration and connectivity
# Usage: bash validate-redis.sh [redis_host] [redis_port]

set -e

REDIS_HOST="${1:-localhost}"
REDIS_PORT="${2:-6379}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== Redis Configuration Validation ==="
echo ""

# Check if redis-cli is installed
if ! command -v redis-cli &> /dev/null; then
    echo -e "${RED}✗ redis-cli not found${NC}"
    echo "  Install with: brew install redis (macOS) or apt-get install redis-tools (Linux)"
    exit 1
fi
echo -e "${GREEN}✓ redis-cli found${NC}"

# Test connection
echo ""
echo "Testing connection to Redis at ${REDIS_HOST}:${REDIS_PORT}..."
if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis connection successful${NC}"
else
    echo -e "${RED}✗ Cannot connect to Redis at ${REDIS_HOST}:${REDIS_PORT}${NC}"
    exit 1
fi

# Get Redis info
echo ""
echo "Redis server information:"
REDIS_VERSION=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info server | grep redis_version | cut -d: -f2 | tr -d '\r')
REDIS_MODE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info server | grep redis_mode | cut -d: -f2 | tr -d '\r')
echo -e "  Version: ${GREEN}${REDIS_VERSION}${NC}"
echo -e "  Mode: ${GREEN}${REDIS_MODE}${NC}"

# Check memory configuration
echo ""
echo "Memory configuration:"
MAX_MEMORY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get maxmemory | tail -1)
MAX_MEMORY_POLICY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get maxmemory-policy | tail -1)
USED_MEMORY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
echo -e "  Max Memory: ${GREEN}${MAX_MEMORY}${NC}"
echo -e "  Eviction Policy: ${GREEN}${MAX_MEMORY_POLICY}${NC}"
echo -e "  Current Usage: ${GREEN}${USED_MEMORY}${NC}"

# Check persistence configuration
echo ""
echo "Persistence configuration:"
SAVE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get save | tail -1)
APPENDONLY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get appendonly | tail -1)
echo -e "  RDB Save: ${GREEN}${SAVE}${NC}"
echo -e "  AOF Enabled: ${GREEN}${APPENDONLY}${NC}"

# Check connection pool configuration (recommended for Go)
echo ""
echo -e "${YELLOW}Recommended Go Redis client configuration:${NC}"
cat <<EOF
redis.Options{
    Addr:         "${REDIS_HOST}:${REDIS_PORT}",
    MaxRetries:   3,
    PoolSize:     10,
    MinIdleConns: 5,
    MaxConnAge:   time.Hour,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}
EOF

# Test basic operations
echo ""
echo "Testing basic operations..."
TEST_KEY="test:validation:$(date +%s)"
TEST_VALUE="platform-go-cache-test"

if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" SET "$TEST_KEY" "$TEST_VALUE" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ SET operation successful${NC}"
else
    echo -e "${RED}✗ SET operation failed${NC}"
    exit 1
fi

RETRIEVED=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" GET "$TEST_KEY")
if [ "$RETRIEVED" = "$TEST_VALUE" ]; then
    echo -e "${GREEN}✓ GET operation successful${NC}"
else
    echo -e "${RED}✗ GET operation failed${NC}"
    exit 1
fi

# Cleanup
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DEL "$TEST_KEY" > /dev/null 2>&1
echo -e "${GREEN}✓ DEL operation successful${NC}"

echo ""
echo -e "${GREEN}=== All validation checks passed ===${NC}"
