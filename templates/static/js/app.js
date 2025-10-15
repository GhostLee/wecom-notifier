let base64Image = '';

// 图片文件选择处理
document.getElementById('imageFile').addEventListener('change', function(e) {
    const file = e.target.files[0];
    if (file) {
        // 检查文件大小 (限制 2MB)
        if (file.size > 2 * 1024 * 1024) {
            showResult('图片大小不能超过 2MB', false);
            return;
        }

        const reader = new FileReader();
        reader.onload = function(e) {
            base64Image = e.target.result.split(',')[1];
            document.getElementById('imagePreview').src = e.target.result;
            document.getElementById('imagePreview').style.display = 'block';
        };
        reader.readAsDataURL(file);
    }
});

// 发送消息
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

    // 禁用所有按钮
    const buttons = document.querySelectorAll('button');
    buttons.forEach(btn => btn.disabled = true);

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
            showResult('✅ 消息发送成功！', true);
        } else {
            showResult('❌ 发送失败: ' + (data.error || JSON.stringify(data)), false);
        }
    } catch (error) {
        showResult('❌ 请求失败: ' + error.message, false);
    } finally {
        // 恢复按钮状态
        buttons.forEach(btn => btn.disabled = false);
    }
}

// 显示结果消息
function showResult(message, success) {
    const resultDiv = document.getElementById('result');
    resultDiv.textContent = message;
    resultDiv.className = 'result ' + (success ? 'success' : 'error');
    resultDiv.style.display = 'block';

    setTimeout(() => {
        resultDiv.style.display = 'none';
    }, 5000);
}

// 键盘快捷键支持
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + Enter 发送文本消息
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        const content = document.getElementById('content').value;
        if (content) {
            sendMessage('text');
        }
    }
});

// 保存 API Key 到 sessionStorage
const apiKeyInput = document.getElementById('apiKey');
const savedApiKey = sessionStorage.getItem('apiKey');
if (savedApiKey) {
    apiKeyInput.value = savedApiKey;
}

apiKeyInput.addEventListener('change', function() {
    sessionStorage.setItem('apiKey', this.value);
});