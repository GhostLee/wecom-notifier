#!/usr/bin/env python3
"""
企业微信通知服务 Python 客户端示例
"""

import requests
import base64
from typing import Optional


class WeComNotifier:
    """企业微信通知客户端"""

    def __init__(self, api_url: str, api_key: str):
        """
        初始化客户端

        Args:
            api_url: API 地址，如 http://localhost:8080
            api_key: API 密钥
        """
        self.api_url = api_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        })

    def send_text(self, text: str, touser: str = '@all') -> dict:
        """
        发送文本消息

        Args:
            text: 消息内容
            touser: 接收人，默认 @all

        Returns:
            API 响应
        """
        url = f'{self.api_url}/api/send/text'
        data = {
            'text': text,
            'touser': touser
        }

        response = self.session.post(url, json=data)
        return response.json()

    def send_image(self, image_path: str, touser: str = '@all') -> dict:
        """
        发送图片消息

        Args:
            image_path: 图片文件路径
            touser: 接收人，默认 @all

        Returns:
            API 响应
        """
        with open(image_path, 'rb') as f:
            image_data = f.read()

        base64_image = base64.b64encode(image_data).decode('utf-8')

        url = f'{self.api_url}/api/send/image'
        data = {
            'image': base64_image,
            'touser': touser
        }

        response = self.session.post(url, json=data)
        return response.json()

    def send_markdown(self, markdown: str, touser: str = '@all') -> dict:
        """
        发送 Markdown 消息

        Args:
            markdown: Markdown 内容
            touser: 接收人，默认 @all

        Returns:
            API 响应
        """
        url = f'{self.api_url}/api/send/markdown'
        data = {
            'markdown': markdown,
            'touser': touser
        }

        response = self.session.post(url, json=data)
        return response.json()

    def health_check(self) -> dict:
        """
        健康检查

        Returns:
            健康状态
        """
        url = f'{self.api_url}/api/health'
        response = self.session.get(url)
        return response.json()


def main():
    """示例使用"""

    # 初始化客户端
    notifier = WeComNotifier(
        api_url='http://localhost:8080',
        api_key='your_api_key'
    )

    # 健康检查
    print('健康检查...')
    result = notifier.health_check()
    print(f'结果: {result}')
    print()

    # 发送文本消息
    print('发送文本消息...')
    result = notifier.send_text('这是通过 Python 客户端发送的测试消息')
    print(f'结果: {result}')
    print()

    # 发送 Markdown 消息
    print('发送 Markdown 消息...')
    markdown = """
# Python 客户端测试

这是通过 **Python 客户端** 发送的 Markdown 消息

## 功能特性

- ✅ 文本消息
- ✅ 图片消息
- ✅ Markdown 消息

## 代码示例

```python
notifier.send_text('Hello World')
```
"""
    result = notifier.send_markdown(markdown)
    print(f'结果: {result}')
    print()

    # 发送图片消息（如果有图片文件）
    # print('发送图片消息...')
    # result = notifier.send_image('path/to/image.jpg')
    # print(f'结果: {result}')

    print('测试完成！')


if __name__ == '__main__':
    main()