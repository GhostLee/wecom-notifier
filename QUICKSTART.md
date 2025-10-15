# 快速启动指南

## 方式一：一键安装脚本（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/GhostLee/wecom-notifier/main/install.sh | bash
```

然后编辑配置文件：

```bash
vim ~/.config/wecom-notifier/.env
```

启动服务：

```bash
wecom-notifier
```

## 方式二：使用 Docker

```bash
# 1. 创建配置文件
cat > .env << EOF
WECOM_CORP_ID=your_corp_id
WECOM_AGENT_ID=1000002
WECOM_SECRET=your_secret
API_KEY=your_secure_key
EOF

# 2. 运行容器
docker run -d \
  --name wecom-notifier \
  -p 8080:8080 \
  --env-file .env \
  abellee000/wecom-notifier:latest

# 3. 查看日志
docker logs -f wecom-notifier
```

## 方式三：使用 Docker Compose

```bash
# 1. 克隆仓库
git clone https://github.com/GhostLee/wecom-notifier.git
cd wecom-notifier

# 2. 配置环境变量
cp .env.example .env
vim .env

# 3. 启动服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f
```

## 方式四：下载预编译二进制

### Linux/macOS

```bash
# 1. 下载对应架构的文件
# 从 https://github.com/GhostLee/wecom-notifier/releases 下载

# 2. 解压
tar -xzf wecom-notifier-linux-amd64.tar.gz

# 3. 添加执行权限
chmod +x wecom-notifier-linux-amd64

# 4. 创建配置文件
cat > .env << EOF
WECOM_CORP_ID=your_corp_id
WECOM_AGENT_ID=1000002
WECOM_SECRET=your_secret
API_KEY=your_secure_key
EOF

# 5. 运行
./wecom-notifier-linux-amd64
```

### Windows

```powershell
# 1. 从 Releases 页面下载 wecom-notifier-windows-amd64.exe.tar.gz

# 2. 解压文件

# 3. 创建 .env 文件在同一目录

# 4. 双击运行或命令行运行
.\wecom-notifier-windows-amd64.exe
```

## 方式五：从源码编译

```bash
# 1. 克隆仓库
git clone https://github.com/GhostLee/wecom-notifier.git
cd wecom-notifier

# 2. 安装依赖
go mod download

# 3. 配置环境变量
cp .env.example .env
vim .env

# 4. 编译运行
make run

# 或直接编译
go build -o wecom-notifier .
./wecom-notifier
```

## 获取企业微信配置

### 1. 获取企业 ID (WECOM_CORP_ID)

1. 登录[企业微信管理后台](https://work.weixin.qq.com/)
2. 我的企业 → 企业信息 → 企业ID

### 2. 创建应用获取 AgentId 和 Secret

1. 应用管理 → 自建 → 创建应用
2. 记录 AgentId (WECOM_AGENT_ID)
3. 查看 Secret (WECOM_SECRET)

### 3. 配置应用可见范围

在应用详情中设置可见范围，选择需要接收消息的成员。

## 测试服务

### 1. 访问测试页面

打开浏览器访问：`http://localhost:8080`

### 2. 使用 API 测试

```bash
# 发送文本消息
curl -X POST http://localhost:8080/api/send/text \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "这是一条测试消息",
    "touser": "@all"
  }'

# 发送 Markdown 消息
curl -X POST http://localhost:8080/api/send/markdown \
  -H "X-API-Key: your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "markdown": "# 标题\n这是**粗体**文本",
    "touser": "@all"
  }'
```

## 常见问题

### Q: 服务启动失败

**A:** 检查以下几点：
1. 配置文件是否正确
2. 端口 8080 是否被占用
3. 查看日志文件排查错误

### Q: 消息发送失败

**A:** 可能的原因：
1. 企业微信配置错误
2. API Key 不正确
3. 接收人不在可见范围
4. 网络连接问题

### Q: Access Token 获取失败

**A:** 检查：
1. WECOM_CORP_ID 是否正确
2. WECOM_SECRET 是否正确
3. 应用是否被停用

## 下一步

- 阅读完整[文档](README.md)
- 配置 [MQTT 集成](README.md#mqtt-使用)
- 部署为[系统服务](README.md#docker-部署)
- 查看 [API 文档](README.md#api-接口)

## 获取帮助

- 提交 [Issue](https://github.com/GhostLee/wecom-notifier/issues)
- 查看 [Wiki](https://github.com/GhostLee/wecom-notifier/wiki)
- 加入讨论 [Discussions](https://github.com/GhostLee/wecom-notifier/discussions)