#!/bin/bash
# Analyze and report cache metrics and performance
# Usage: bash cache-metrics.sh [redis_host] [redis_port]

REDIS_HOST="${1:-localhost}"
REDIS_PORT="${2:-6379}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check redis-cli availability
if ! command -v redis-cli &> /dev/null; then
    echo -e "${RED}redis-cli not found${NC}"
    exit 1
fi

# Check connection
if ! redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null 2>&1; then
    echo -e "${RED}Cannot connect to Redis at ${REDIS_HOST}:${REDIS_PORT}${NC}"
    exit 1
fi

echo "=== Redis Cache Metrics Report ==="
echo "Target: ${REDIS_HOST}:${REDIS_PORT}"
echo "Timestamp: $(date)"
echo ""

# Memory Statistics
echo -e "${BLUE}Memory Statistics:${NC}"
MEMORY_INFO=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info memory)
echo "$MEMORY_INFO" | grep -E "used_memory|used_memory_peak|total_system_memory|maxmemory" | sed 's/^/  /'
echo ""

# Stats (Cache hits/misses)
echo -e "${BLUE}Cache Hit/Miss Statistics:${NC}"
STATS_INFO=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info stats)
TOTAL_COMMANDS=$(echo "$STATS_INFO" | grep "total_commands_processed" | cut -d: -f2 | tr -d '\r')
TOTAL_CONNECTIONS=$(echo "$STATS_INFO" | grep "total_connections_received" | cut -d: -f2 | tr -d '\r')

echo "  Total Commands: ${TOTAL_COMMANDS}"
echo "  Total Connections: ${TOTAL_CONNECTIONS}"
echo ""

# Keyspace statistics
echo -e "${BLUE}Keyspace Statistics:${NC}"
KEYSPACE_INFO=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info keyspace)
if [ -z "$(echo "$KEYSPACE_INFO" | grep -v "^#")" ]; then
    echo "  No keys in Redis"
else
    echo "$KEYSPACE_INFO" | grep -v "^#" | sed 's/^/  /'
fi
echo ""

# List keys by pattern (cache statistics)
echo -e "${BLUE}Cache Keys by Pattern:${NC}"
PATTERNS=("cache:*" "lock:*" "session:*" "group:*")

for pattern in "${PATTERNS[@]}"; do
    COUNT=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" keys "$pattern" | wc -l)
    if [ "$COUNT" -gt 0 ]; then
        echo -e "  ${pattern}: ${GREEN}${COUNT} keys${NC}"
    fi
done
echo ""

# Connection information
echo -e "${BLUE}Connection Pool Information:${NC}"
CONNECTED_CLIENTS=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info clients | grep connected_clients | cut -d: -f2 | tr -d '\r')
echo "  Connected Clients: ${CONNECTED_CLIENTS}"
echo ""

# Check for evictions (might indicate memory pressure)
echo -e "${BLUE}Memory Pressure Check:${NC}"
EVICTED=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info stats | grep evicted_keys | cut -d: -f2 | tr -d '\r')
if [ "$EVICTED" -gt 0 ]; then
    echo -e "  ${YELLOW}Warning: ${EVICTED} keys have been evicted${NC}"
    echo "  Consider increasing maxmemory or optimizing cache keys"
else
    echo -e "  ${GREEN}No keys evicted (memory usage healthy)${NC}"
fi
echo ""

# Replication status
echo -e "${BLUE}Replication Status:${NC}"
ROLE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info replication | grep "^role:" | cut -d: -f2 | tr -d '\r')
echo "  Role: ${ROLE}"
echo ""

# Performance recommendations
echo -e "${YELLOW}=== Performance Recommendations ===${NC}"

MAX_MEMORY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get maxmemory | tail -1)
if [ "$MAX_MEMORY" = "0" ]; then
    echo -e "${YELLOW}⚠ No maxmemory limit set${NC}"
    echo "  Set with: redis-cli CONFIG SET maxmemory 512mb"
fi

PERSISTENCE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get save | tail -1)
if [ "$PERSISTENCE" = "" ] || [ "$PERSISTENCE" = "(empty array)" ]; then
    echo -e "${YELLOW}⚠ RDB persistence disabled${NC}"
    echo "  Consider enabling for durability"
fi

APPENDONLY=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" config get appendonly | tail -1)
if [ "$APPENDONLY" != "yes" ]; then
    echo -e "${YELLOW}⚠ AOF persistence disabled${NC}"
    echo "  Consider enabling for better durability: CONFIG SET appendonly yes"
fi

echo ""
echo -e "${GREEN}Report generated successfully${NC}"
