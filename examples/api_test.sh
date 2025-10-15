#!/bin/bash

# API 测试脚本

# 配置
API_URL="${API_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-your_api_key}"

# 颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}企业微信通知 API 测试脚本${NC}"
echo "API URL: ${API_URL}"
echo ""

# 测试函数
test_api() {
    local test_name=$1
    local endpoint=$2
    local data=$3

    echo -e "${GREEN}[TEST]${NC} ${test_name}"

    response=$(curl -s -w "\n%{http_code}" -X POST \
        "${API_URL}${endpoint}" \
        -H "X-API-Key: ${API_KEY}" \
        -H "Content-Type: application/json" \
        -d "${data}")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓${NC} HTTP ${http_code}"
        echo "响应: ${body}"
    else
        echo -e "${RED}✗${NC} HTTP ${http_code}"
        echo "响应: ${body}"
    fi
    echo ""
}

# 测试 1: 健康检查
echo -e "${GREEN}[TEST]${NC} 健康检查"
response=$(curl -s -w "\n%{http_code}" -X GET \
    "${API_URL}/api/health" \
    -H "X-API-Key: ${API_KEY}")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓${NC} HTTP ${http_code}"
    echo "响应: ${body}"
else
    echo -e "${RED}✗${NC} HTTP ${http_code}"
    echo "响应: ${body}"
fi
echo ""

sleep 1

# 测试 2: 发送文本消息
test_api "发送文本消息" "/api/send/text" '{
    "text": "这是通过 API 发送的测试消息",
    "touser": "@all"
}'

sleep 2

# 测试 3: 发送 Markdown 消息
test_api "发送 Markdown 消息" "/api/send/markdown" '{
    "markdown": "# API 测试\n\n这是**Markdown**格式的消息\n\n## 功能列表\n- ✅ 文本消息\n- ✅ Markdown 消息\n- ✅ 图片消息",
    "touser": "@all"
}'

sleep 2

# 测试 4: 发送给特定用户
test_api "发送给特定用户" "/api/send/text" '{
    "text": "这是发送给特定用户的消息",
    "touser": "UserID"
}'

sleep 2

# 测试 5: 测试错误处理 - 缺少参数
echo -e "${GREEN}[TEST]${NC} 错误处理 - 缺少参数"
response=$(curl -s -w "\n%{http_code}" -X POST \
    "${API_URL}/api/send/text" \
    -H "X-API-Key: ${API_KEY}" \
    -H "Content-Type: application/json" \
    -d '{}')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "400" ]; then
    echo -e "${GREEN}✓${NC} HTTP ${http_code} (预期错误)"
    echo "响应: ${body}"
else
    echo -e "${RED}✗${NC} HTTP ${http_code}"
    echo "响应: ${body}"
fi
echo ""

# 测试 6: 测试错误处理 - 错误的 API Key
echo -e "${GREEN}[TEST]${NC} 错误处理 - 错误的 API Key"
response=$(curl -s -w "\n%{http_code}" -X POST \
    "${API_URL}/api/send/text" \
    -H "X-API-Key: wrong_key" \
    -H "Content-Type: application/json" \
    -d '{"text":"test"}')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓${NC} HTTP ${http_code} (预期错误)"
    echo "响应: ${body}"
else
    echo -e "${RED}✗${NC} HTTP ${http_code}"
    echo "响应: ${body}"
fi
echo ""

echo -e "${BLUE}测试完成！${NC}"