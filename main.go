package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gookit/slog/handler"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
	"github.com/joho/godotenv"
)

// WeComConfig 企业微信配置
type WeComConfig struct {
	CorpID      string
	AgentID     int
	Secret      string
	ToUser      string
	AccessToken string
	ExpiresAt   time.Time
	mu          sync.RWMutex
}

// TokenResponse 获取token响应
type TokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// UploadResponse 上传文件响应
type UploadResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

// MessageResponse 发送消息响应
type MessageResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"invalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// SendTextRequest 发送文本请求
type SendTextRequest struct {
	Text   string `json:"text" binding:"required"`
	ToUser string `json:"touser"`
}

// SendImageRequest 发送图片请求
type SendImageRequest struct {
	Image  string `json:"image" binding:"required"` // base64编码
	ToUser string `json:"touser"`
}

// SendMarkdownRequest 发送Markdown请求
type SendMarkdownRequest struct {
	Markdown string `json:"markdown" binding:"required"`
	ToUser   string `json:"touser"`
}

// MQTTMessage MQTT消息格式
type MQTTMessage struct {
	Type    string `json:"type"` // text, image, markdown
	Content string `json:"content"`
	ToUser  string `json:"touser"`
}

var (
	wecomConfig *WeComConfig
	apiKey      string
	mqttClient  mqtt.Client
)

func init() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using system environment variables")
	}

	// 初始化日志
	initLogger()

	// 初始化企业微信配置
	wecomConfig = &WeComConfig{
		CorpID: os.Getenv("WECOM_CORP_ID"),
		Secret: os.Getenv("WECOM_SECRET"),
		ToUser: getEnvOrDefault("WECOM_TO_USER", "@all"),
	}

	agentID, err := strconv.Atoi(os.Getenv("WECOM_AGENT_ID"))
	if err != nil {
		slog.Fatal("Invalid WECOM_AGENT_ID")
	}
	wecomConfig.AgentID = agentID

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		slog.Fatal("API_KEY is required")
	}

	slog.Info("Configuration loaded successfully")
}

func initLogger() {
	// 获取日志配置
	logLevel := getEnvOrDefault("LOG_LEVEL", "info")
	logDir := getEnvOrDefault("LOG_DIR", "./logs")
	logMaxAge, _ := strconv.Atoi(getEnvOrDefault("LOG_MAX_AGE_DAYS", "30"))
	logRotate := getEnvOrDefault("LOG_ROTATE", "true") == "true"

	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	// 配置日志处理器
	handlers := []slog.Handler{
		handler.NewConsoleHandler(slog.AllLevels),
	}

	if logRotate {
		logFile := filepath.Join(logDir, "wecom-notifier.log")
		fileHandler, err := handler.JSONFileHandler(logFile)
		if err != nil {
			fmt.Printf("Failed to create file handler: %v\n", err)
			os.Exit(1)
		}
		handlers = append(handlers, fileHandler)

		// 启动日志清理协程
		go cleanOldLogs(logDir, logMaxAge)
	}

	slog.Configure(func(logger *slog.SugaredLogger) {
		logger.ReportCaller = true
		logger.SetHandlers(handlers)
		logger.Level = slog.LevelByName(logLevel)
	})
}

func cleanOldLogs(logDir string, maxAgeDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		files, err := os.ReadDir(logDir)
		if err != nil {
			slog.Errorf("Failed to read log directory: %v", err)
			continue
		}

		cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				logPath := filepath.Join(logDir, file.Name())
				if err := os.Remove(logPath); err != nil {
					slog.Errorf("Failed to remove old log file %s: %v", logPath, err)
				} else {
					slog.Infof("Removed old log file: %s", logPath)
				}
			}
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 获取并缓存 access_token
func (w *WeComConfig) GetAccessToken() (string, error) {
	w.mu.RLock()
	if w.AccessToken != "" && time.Now().Before(w.ExpiresAt) {
		token := w.AccessToken
		w.mu.RUnlock()
		return token, nil
	}
	w.mu.RUnlock()

	w.mu.Lock()
	defer w.mu.Unlock()

	// 双重检查
	if w.AccessToken != "" && time.Now().Before(w.ExpiresAt) {
		return w.AccessToken, nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		w.CorpID, w.Secret)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("get token failed: %s", tokenResp.ErrMsg)
	}

	w.AccessToken = tokenResp.AccessToken
	w.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 提前5分钟过期

	slog.Infof("Access token refreshed, expires at: %s", w.ExpiresAt.Format(time.RFC3339))
	return w.AccessToken, nil
}

// 定期刷新 access_token
func startTokenRefresher(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := wecomConfig.GetAccessToken(); err != nil {
					slog.Errorf("Failed to refresh access token: %v", err)
				}
			}
		}
	}()
}

// 发送文本消息
func sendTextMessage(text, toUser string) (*MessageResponse, error) {
	token, err := wecomConfig.GetAccessToken()
	if err != nil {
		return nil, err
	}

	if toUser == "" {
		toUser = wecomConfig.ToUser
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	data := map[string]interface{}{
		"touser":  toUser,
		"agentid": wecomConfig.AgentID,
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
		"duplicate_check_interval": 600,
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &msgResp, nil
}

// 发送图片消息
func sendImageMessage(base64Content, toUser string) (*MessageResponse, error) {
	token, err := wecomConfig.GetAccessToken()
	if err != nil {
		return nil, err
	}

	if toUser == "" {
		toUser = wecomConfig.ToUser
	}

	// 解码 base64
	imageData, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// 上传图片
	uploadURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=image", token)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", "image.jpg")
	if err != nil {
		return nil, err
	}
	part.Write(imageData)
	writer.Close()

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}
	defer resp.Body.Close()

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	if uploadResp.ErrCode != 0 {
		return nil, fmt.Errorf("upload failed: %s", uploadResp.ErrMsg)
	}

	// 发送消息
	sendURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	data := map[string]interface{}{
		"touser":  toUser,
		"agentid": wecomConfig.AgentID,
		"msgtype": "image",
		"image": map[string]string{
			"media_id": uploadResp.MediaID,
		},
		"duplicate_check_interval": 600,
	}

	jsonData, _ := json.Marshal(data)
	resp2, err := http.Post(sendURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp2.Body.Close()

	var msgResp MessageResponse
	if err := json.NewDecoder(resp2.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &msgResp, nil
}

// 发送 Markdown 消息
func sendMarkdownMessage(markdown, toUser string) (*MessageResponse, error) {
	token, err := wecomConfig.GetAccessToken()
	if err != nil {
		return nil, err
	}

	if toUser == "" {
		toUser = wecomConfig.ToUser
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	data := map[string]interface{}{
		"touser":  toUser,
		"agentid": wecomConfig.AgentID,
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": markdown,
		},
		"duplicate_check_interval": 600,
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &msgResp, nil
}

// API Key 中间件
func apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}

		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 初始化 MQTT 客户端
func initMQTT() {
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		slog.Info("MQTT_BROKER not configured, skipping MQTT initialization")
		return
	}

	mqttClientID := getEnvOrDefault("MQTT_CLIENT_ID", "wecom-notifier")
	mqttTopic := getEnvOrDefault("MQTT_TOPIC", "wecom/notify")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(mqttClientID)
	if mqttUsername != "" {
		opts.SetUsername(mqttUsername)
		opts.SetPassword(mqttPassword)
	}
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		slog.Info("Connected to MQTT broker")
		if token := client.Subscribe(mqttTopic, 0, onMQTTMessage); token.Wait() && token.Error() != nil {
			slog.Errorf("Failed to subscribe to topic: %v", token.Error())
		} else {
			slog.Infof("Subscribed to topic: %s", mqttTopic)
		}
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		slog.Errorf("MQTT connection lost: %v", err)
	})

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		slog.Errorf("Failed to connect to MQTT broker: %v", token.Error())
	}
}

// MQTT 消息处理
func onMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	slog.Infof("Received MQTT message from topic: %s", msg.Topic())

	var mqttMsg MQTTMessage
	if err := json.Unmarshal(msg.Payload(), &mqttMsg); err != nil {
		slog.Errorf("Failed to parse MQTT message: %v", err)
		return
	}

	var err error
	var resp *MessageResponse

	switch mqttMsg.Type {
	case "text":
		resp, err = sendTextMessage(mqttMsg.Content, mqttMsg.ToUser)
	case "image":
		resp, err = sendImageMessage(mqttMsg.Content, mqttMsg.ToUser)
	case "markdown":
		resp, err = sendMarkdownMessage(mqttMsg.Content, mqttMsg.ToUser)
	default:
		slog.Errorf("Unknown message type: %s", mqttMsg.Type)
		return
	}

	if err != nil {
		slog.Errorf("Failed to send message via MQTT: %v", err)
		return
	}

	if resp.ErrCode != 0 {
		slog.Errorf("Send message failed: %s", resp.ErrMsg)
	} else {
		slog.Info("Message sent successfully via MQTT")
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 token 刷新协程
	startTokenRefresher(ctx)

	// 初始化 MQTT
	initMQTT()

	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 首页 - 测试页面
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, getTestPageHTML())
	})

	// API 路由组
	api := r.Group("/api")
	api.Use(apiKeyAuth())
	{
		// 发送文本消息
		api.POST("/send/text", func(c *gin.Context) {
			var req SendTextRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := sendTextMessage(req.Text, req.ToUser)
			if err != nil {
				slog.Errorf("Failed to send text message: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if resp.ErrCode != 0 {
				c.JSON(http.StatusOK, gin.H{"success": false, "error": resp.ErrMsg, "response": resp})
				return
			}

			slog.Info("Text message sent successfully")
			c.JSON(http.StatusOK, gin.H{"success": true, "response": resp})
		})

		// 发送图片消息
		api.POST("/send/image", func(c *gin.Context) {
			var req SendImageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := sendImageMessage(req.Image, req.ToUser)
			if err != nil {
				slog.Errorf("Failed to send image message: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if resp.ErrCode != 0 {
				c.JSON(http.StatusOK, gin.H{"success": false, "error": resp.ErrMsg, "response": resp})
				return
			}

			slog.Info("Image message sent successfully")
			c.JSON(http.StatusOK, gin.H{"success": true, "response": resp})
		})

		// 发送 Markdown 消息
		api.POST("/send/markdown", func(c *gin.Context) {
			var req SendMarkdownRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := sendMarkdownMessage(req.Markdown, req.ToUser)
			if err != nil {
				slog.Errorf("Failed to send markdown message: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if resp.ErrCode != 0 {
				c.JSON(http.StatusOK, gin.H{"success": false, "error": resp.ErrMsg, "response": resp})
				return
			}

			slog.Info("Markdown message sent successfully")
			c.JSON(http.StatusOK, gin.H{"success": true, "response": resp})
		})

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
	}

	port := getEnvOrDefault("PORT", "8080")
	slog.Infof("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		slog.Fatalf("Failed to start server: %v", err)
	}
}

func getTestPageHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>企业微信通知测试</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; padding: 20px; }
        .container { max-width: 800px; margin: 0 auto; background: white; border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); overflow: hidden; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { font-size: 28px; margin-bottom: 8px; }
        .header p { opacity: 0.9; font-size: 14px; }
        .content { padding: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 8px; font-weight: 600; color: #333; font-size: 14px; }
        input, textarea { width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 8px; font-size: 14px; transition: border-color 0.3s; }
        input:focus, textarea:focus { outline: none; border-color: #667eea; }
        textarea { min-height: 120px; resize: vertical; font-family: inherit; }
        .btn-group { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-top: 24px; }
        button { padding: 14px; border: none; border-radius: 8px; font-size: 15px; font-weight: 600; cursor: pointer; transition: all 0.3s; }
        .btn-text { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .btn-image { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; }
        .btn-markdown { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; }
        button:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(0,0,0,0.2); }
        button:active { transform: translateY(0); }
        .result { margin-top: 24px; padding: 16px; border-radius: 8px; display: none; }
        .result.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .result.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .file-input-wrapper { position: relative; overflow: hidden; display: inline-block; width: 100%; }
        .file-input-wrapper input[type=file] { position: absolute; left: -9999px; }
        .file-input-label { display: block; padding: 12px; background: #f8f9fa; border: 2px dashed #667eea; border-radius: 8px; text-align: center; cursor: pointer; transition: all 0.3s; }
        .file-input-label:hover { background: #e9ecef; border-color: #764ba2; }
        .preview { margin-top: 12px; max-width: 200px; border-radius: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📱 企业微信通知测试</h1>
            <p>测试发送文本、图片和 Markdown 消息</p>
        </div>
        <div class="content">
            <div class="form-group">
                <label>API Key</label>
                <input type="password" id="apiKey" placeholder="请输入 API Key">
            </div>
            <div class="form-group">
                <label>接收人 (可选，默认 @all)</label>
                <input type="text" id="toUser" placeholder="@all">
            </div>
            <div class="form-group">
                <label>消息内容</label>
                <textarea id="content" placeholder="输入文本消息或 Markdown 内容"></textarea>
            </div>
            <div class="form-group">
                <label>图片 (用于图片消息)</label>
                <div class="file-input-wrapper">
                    <input type="file" id="imageFile" accept="image/*">
                    <label for="imageFile" class="file-input-label">点击选择图片</label>
                </div>
                <img id="imagePreview" class="preview" style="display:none;">
            </div>
            <div class="btn-group">
                <button class="btn-text" onclick="sendMessage('text')">发送文本</button>
                <button class="btn-image" onclick="sendMessage('image')">发送图片</button>
                <button class="btn-markdown" onclick="sendMessage('markdown')">发送 Markdown</button>
            </div>
            <div id="result" class="result"></div>
        </div>
    </div>

    <script>
        let base64Image = '';
        
        document.getElementById('imageFile').addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    base64Image = e.target.result.split(',')[1];
                    document.getElementById('imagePreview').src = e.target.result;
                    document.getElementById('imagePreview').style.display = 'block';
                };
                reader.readAsDataURL(file);
            }
        });

        async function sendMessage(type) {
            const apiKey = document.getElementById('apiKey').value;
            const toUser = document.getElementById('toUser').value;
            const content = document.getElementById('content').value;
            const resultDiv = document.getElementById('result');

            if (!apiKey) {
                showResult('请输入 API Key', false);
                return;
            }

            let endpoint = '';
            let body = { touser: toUser || '@all' };

            if (type === 'text') {
                if (!content) {
                    showResult('请输入文本内容', false);
                    return;
                }
                endpoint = '/api/send/text';
                body.text = content;
            } else if (type === 'image') {
                if (!base64Image) {
                    showResult('请选择图片', false);
                    return;
                }
                endpoint = '/api/send/image';
                body.image = base64Image;
            } else if (type === 'markdown') {
                if (!content) {
                    showResult('请输入 Markdown 内容', false);
                    return;
                }
                endpoint = '/api/send/markdown';
                body.markdown = content;
            }

            try {
                const response = await fetch(endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-API-Key': apiKey
                    },
                    body: JSON.stringify(body)
                });

                const data = await response.json();
                
                if (response.ok && data.success) {
                    showResult('消息发送成功！', true);
                } else {
                    showResult('发送失败: ' + (data.error || JSON.stringify(data)), false);
                }
            } catch (error) {
                showResult('请求失败: ' + error.message, false);
            }
        }

        function showResult(message, success) {
            const resultDiv = document.getElementById('result');
            resultDiv.textContent = message;
            resultDiv.className = 'result ' + (success ? 'success' : 'error');
            resultDiv.style.display = 'block';
            setTimeout(() => {
                resultDiv.style.display = 'none';
            }, 5000);
        }
    </script>
</body>
</html>`
}
