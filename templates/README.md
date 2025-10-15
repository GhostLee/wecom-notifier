# 模板文件说明

本目录包含 Web 测试界面的所有前端资源。

## 目录结构

```
templates/
├── README.md              # 本文件
├── index.html            # 主页面模板
└── static/               # 静态资源目录
    ├── css/
    │   └── style.css     # 样式文件
    └── js/
        └── app.js        # JavaScript 逻辑
```

## 文件说明

### index.html

主测试页面，提供用户界面用于：
- 输入 API Key 进行身份验证
- 选择接收人（默认 @all）
- 输入消息内容（文本/Markdown）
- 上传图片文件
- 发送三种类型的消息

### static/css/style.css

页面样式表，包含：
- 响应式布局设计
- 渐变背景和卡片样式
- 表单元素样式
- 按钮动画效果
- 结果提示样式
- 移动端适配

### static/js/app.js

前端交互逻辑，实现：
- 图片文件读取和 Base64 编码
- 表单验证
- API 请求发送
- 结果展示
- API Key 会话存储
- 键盘快捷键支持（Ctrl/Cmd + Enter）

## 自定义修改

### 修改页面标题

编辑 `index.html` 中的 `<title>` 和 `<h1>` 标签。

### 修改主题颜色

编辑 `static/css/style.css` 中的渐变色：

```css
/* 主背景渐变 */
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

/* 按钮渐变 */
.btn-text {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
```

### 添加新功能

在 `static/js/app.js` 中添加新的函数，并在 `index.html` 中调用。

## 注意事项

1. **路径引用**: 静态资源使用 `/static/` 前缀
2. **API 路径**: 所有 API 请求使用 `/api/` 前缀
3. **会话存储**: API Key 存储在 sessionStorage，刷新页面不会丢失
4. **文件大小**: 图片上传限制为 2MB

## 开发建议

### 本地开发

修改模板文件后，重启服务即可看到变化：

```bash
make run
```

### 添加新页面

1. 在 `templates/` 目录创建新的 HTML 文件
2. 在 `main.go` 中添加路由：

```go
r.GET("/new-page", func(c *gin.Context) {
    c.HTML(http.StatusOK, "new-page.html", nil)
})
```

### 添加新的静态资源

将文件放入对应目录：
- CSS: `static/css/`
- JavaScript: `static/js/`
- 图片: `static/images/`

在 HTML 中引用：

```html
<link rel="stylesheet" href="/static/css/new-style.css">
<script src="/static/js/new-script.js"></script>
<img src="/static/images/logo.png">
```