# 企业微信通知服务

基于 Go 语言实现的企业微信消息推送服务，支持 HTTP API 和 MQTT 两种方式接收消息请求。

## 功能特性

- ✅ 支持发送文本、图片、Markdown 三种消息类型
- ✅ HTTP API 接口，使用 Gin 框架
- ✅ MQTT 消息订阅支持
- ✅ Access Token 自动获取和刷新机制
- ✅ API Key 认证
- ✅ 完整的日志系统（控制台 + 文件）
- ✅ 日志轮转和自动清理
- ✅ 内置 Web 测试页面
- ✅ 环境变量配置管理

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

复制 `.env.example` 到 `.env` 并修改配置：

```bash
cp .env.example .env
```

必需配置项：
```bash
WECOM_CORP_ID=你的企业ID
WECOM_AGENT_ID=应用的AgentId
WECOM_SECRET=应用的Secret
API_KEY=你的API密钥
```

### 3. 运行程序

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 4. 访问测试页面

打开浏览器访问：`http://localhost:8080`

## API 接口

所有 API 请求需要在 Header 中携带 `X-API-Key` 或在 URL 参数中提供 `api_key`。

### 发送文本消息

```bash
POST /api/send/text
Content-Type: application/json
X-API-Key: your_api_key

{
  "text": "这是一条测试消息",
  "touser": "@all"
}
```

### 发送图片消息

```bash
POST /api/send/image
Content-Type: application/json
X-API-Key: your_api_key

{
  "image": "base64编码的图片数据",
  "touser": "@all"
}
```

### 发送 Markdown 消息

```bash
POST /api/send/markdown
Content-Type: application/json
X-API-Key: your_api_key

{
  "markdown": "# 标题\n这是**粗体**文本",
  "touser": "@all"
}
```

### 健康检查

```bash
GET /api/health
X-API-Key: your_api_key
```

## MQTT 使用

配置 MQTT 相关环境变量后，程序会自动连接到 MQTT Broker 并订阅指定主题。

### 消息格式

发送到 MQTT 主题的消息格式：

```json
{
  "type": "text",
  "content": "消息内容",
  "touser": "@all"
}
```

类型可选：`text`、`image`、`markdown`

### 示例（使用 mosquitto_pub）

```bash
# 发送文本消息
mosquitto_pub -h localhost -t wecom/notify -m '{"type":"text","content":"测试消息","touser":"@all"}'

# 发送 Markdown 消息
mosquitto_pub -h localhost -t wecom/notify -m '{"type":"markdown","content":"# 标题\n内容","touser":"@all"}'
```

## 配置说明

### 企业微信配置

- `WECOM_CORP_ID`: 企业 ID，在企业微信管理后台获取
- `WECOM_AGENT_ID`: 应用的 AgentId
- `WECOM_SECRET`: 应用的 Secret
- `WECOM_TO_USER`: 默认接收人，默认为 `@all`

### 服务配置

- `API_KEY`: API 访问密钥（必需）
- `PORT`: 服务端口，默认 8080
- `GIN_MODE`: Gin 运行模式（debug/release），默认 release

### 日志配置

- `LOG_LEVEL`: 日志级别（trace/debug/info/warn/error），默认 info
- `LOG_DIR`: 日志文件目录，默认 ./logs
- `LOG_MAX_AGE_DAYS`: 日志最大保存天数，默认 30
- `LOG_ROTATE`: 是否启用日志轮转，默认 true

### MQTT 配置（可选）

- `MQTT_BROKER`: MQTT Broker 地址，如 tcp://localhost:1883
- `MQTT_CLIENT_ID`: 客户端 ID，默认 wecom-notifier
- `MQTT_TOPIC`: 订阅的主题，默认 wecom/notify
- `MQTT_USERNAME`: MQTT 用户名（如果需要）
- `MQTT_PASSWORD`: MQTT 密码（如果需要）

## Docker 部署

创建 Dockerfile：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wecom-notifier .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/wecom-notifier .
EXPOSE 8080
CMD ["./wecom-notifier"]
```

构建和运行：

```bash
docker build -t wecom-notifier .
docker run -d \
  --name wecom-notifier \
  -p 8080:8080 \
  --env-file .env \
  wecom-notifier
```

## 注意事项

1. **Access Token 缓存**：程序会自动缓存 Access Token 并在过期前刷新（提前 5 分钟）
2. **消息去重**：所有消息默认 600 秒去重间隔
3. **日志管理**：日志文件会自动轮转，超过保留天数的日志会被自动删除
4. **API 安全**：请妥善保管 API Key，建议使用强密码
5. **图片大小**：上传图片建议不超过 2MB

## 开发调试

启用 debug 模式：

```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
go run main.go
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！