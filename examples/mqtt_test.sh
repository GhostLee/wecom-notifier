#!/bin/bash

# MQTT 测试脚本
# 需要安装 mosquitto-clients: apt-get install mosquitto-clients

# 配置
MQTT_BROKER="${MQTT_BROKER:-localhost}"
MQTT_PORT="${MQTT_PORT:-1883}"
MQTT_TOPIC="${MQTT_TOPIC:-wecom/notify}"
MQTT_USERNAME="${MQTT_USERNAME:-}"
MQTT_PASSWORD="${MQTT_PASSWORD:-}"

# 颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}MQTT 测试脚本${NC}"
echo "Broker: ${MQTT_BROKER}:${MQTT_PORT}"
echo "Topic: ${MQTT_TOPIC}"
echo ""

# 构建认证参数
AUTH_PARAMS=""
if [ -n "$MQTT_USERNAME" ]; then
    AUTH_PARAMS="-u ${MQTT_USERNAME} -P ${MQTT_PASSWORD}"
fi

# 测试 1: 发送文本消息
echo -e "${GREEN}[TEST 1]${NC} 发送文本消息"
mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" $AUTH_PARAMS \
    -t "$MQTT_TOPIC" \
    -m '{
        "type": "text",
        "content": "这是通过 MQTT 发送的测试消息",
        "touser": "@all"
    }'
echo "已发送"
echo ""

sleep 2

# 测试 2: 发送 Markdown 消息
echo -e "${GREEN}[TEST 2]${NC} 发送 Markdown 消息"
mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" $AUTH_PARAMS \
    -t "$MQTT_TOPIC" \
    -m '{
        "type": "markdown",
        "content": "# MQTT 测试\n\n这是**Markdown**格式的消息\n\n- 列表项 1\n- 列表项 2",
        "touser": "@all"
    }'
echo "已发送"
echo ""

sleep 2

# 测试 3: 发送给特定用户
echo -e "${GREEN}[TEST 3]${NC} 发送给特定用户"
mosquitto_pub -h "$MQTT_BROKER" -p "$MQTT_PORT" $AUTH_PARAMS \
    -t "$MQTT_TOPIC" \
    -m '{
        "type": "text",
        "content": "这是发送给特定用户的消息",
        "touser": "UserID"
    }'
echo "已发送"
echo ""

echo "测试完成！请检查企业微信是否收到消息"