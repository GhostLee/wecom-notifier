.PHONY: help build run dev test clean docker-build docker-run docker-stop deps

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  make deps         - 下载依赖"
	@echo "  make build        - 编译程序"
	@echo "  make run          - 运行程序"
	@echo "  make dev          - 开发模式运行"
	@echo "  make test         - 运行测试"
	@echo "  make clean        - 清理编译文件"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make docker-run   - 运行 Docker 容器"
	@echo "  make docker-stop  - 停止 Docker 容器"

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 编译程序
build:
	@echo "编译程序..."
	go build -ldflags="-w -s" -o wecom-notifier .
	@echo "编译完成: wecom-notifier"

# 运行程序
run: build
	@echo "启动服务..."
	./wecom-notifier

# 开发模式运行
dev:
	@echo "开发模式启动..."
	@export GIN_MODE=debug LOG_LEVEL=debug && go run main.go

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 清理
clean:
	@echo "清理编译文件..."
	rm -f wecom-notifier
	rm -rf logs/*
	@echo "清理完成"

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t wecom-notifier:latest .
	@echo "镜像构建完成"

# 运行 Docker 容器
docker-run:
	@echo "启动 Docker 容器..."
	docker-compose up -d
	@echo "容器启动完成"
	@echo "访问: http://localhost:8080"

# 停止 Docker 容器
docker-stop:
	@echo "停止 Docker 容器..."
	docker-compose down
	@echo "容器已停止"

# 查看 Docker 日志
docker-logs:
	docker-compose logs -f wecom-notifier

# 重启 Docker 容器
docker-restart:
	docker-compose restart wecom-notifier

# 安装为 systemd 服务
install-systemd:
	@echo "安装 systemd 服务..."
	@sudo cp wecom-notifier.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable wecom-notifier
	@echo "服务已安装，使用以下命令管理:"
	@echo "  sudo systemctl start wecom-notifier"
	@echo "  sudo systemctl stop wecom-notifier"
	@echo "  sudo systemctl status wecom-notifier"