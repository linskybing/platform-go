#!/bin/bash
# Quick Integration Test Runner
# Usage: ./scripts/quick-test.sh [test-name]

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TEST_NAME="${1:-all}"

echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        Platform-Go 快速整合測試執行器                      ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Export test environment variables
export DATABASE_URL="postgres://testuser:testpass@localhost:5433/platform_test"
export REDIS_URL="redis://localhost:6380/0"
export ENVIRONMENT="test"

cd "$PROJECT_ROOT"

case "$TEST_NAME" in
    all)
        echo -e "${YELLOW}運行所有整合測試...${NC}"
        go test -v -tags=integration -timeout 15m ./test/integration/...
        ;;
    
    user)
        echo -e "${YELLOW}運行用戶相關測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestUser"
        ;;
    
    group)
        echo -e "${YELLOW}運行群組相關測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestGroup"
        ;;
    
    project)
        echo -e "${YELLOW}運行項目相關測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestProject"
        ;;
    
    config)
        echo -e "${YELLOW}運行配置文件測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestConfigFile"
        ;;
    
    audit)
        echo -e "${YELLOW}運行審計日誌測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestAudit"
        ;;
    
    image)
        echo -e "${YELLOW}運行鏡像管理測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestImage"
        ;;
    
    form)
        echo -e "${YELLOW}運行表單管理測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestForm"
        ;;
    
    storage)
        echo -e "${YELLOW}運行存儲權限測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestStorage"
        ;;
    
    pvc)
        echo -e "${YELLOW}運行 PVC 綁定測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestPVC"
        ;;
    
    auth)
        echo -e "${YELLOW}運行認證相關測試...${NC}"
        go test -v -tags=integration -timeout 5m ./test/integration/... -run "TestAuth"
        ;;
    
    k8s)
        echo -e "${YELLOW}運行 Kubernetes 測試...${NC}"
        go test -v -tags=integration -timeout 10m ./test/integration/... -run "TestK8s"
        ;;
    
    list)
        echo -e "${CYAN}可用的測試類型:${NC}"
        echo "  all      - 運行所有測試"
        echo "  user     - 用戶管理測試"
        echo "  group    - 群組管理測試"
        echo "  project  - 項目管理測試"
        echo "  config   - 配置文件測試"
        echo "  audit    - 審計日誌測試"
        echo "  image    - 鏡像管理測試"
        echo "  form     - 表單管理測試"
        echo "  storage  - 存儲權限測試"
        echo "  pvc      - PVC 綁定測試"
        echo "  auth     - 認證相關測試"
        echo "  k8s      - Kubernetes 測試"
        echo ""
        echo "使用方法: $0 [test-type]"
        exit 0
        ;;
    
    *)
        echo -e "${YELLOW}未知的測試類型: $TEST_NAME${NC}"
        echo "使用 '$0 list' 查看可用的測試類型"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✅ 測試執行完成${NC}"
