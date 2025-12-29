.PHONY: build run test clean lint swagger docker-build docker-run help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := youlai-gin
BUILD_DIR := build
MAIN_FILE := main.go

# 编译
build: ## 编译项目
	@echo "正在编译..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✓ 编译完成: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run: ## 运行项目
	@echo "正在启动..."
	go run $(MAIN_FILE)

# 测试
test: ## 运行所有测试
	@echo "正在运行测试..."
	go test -v ./...

# 测试覆盖率
test-coverage: ## 生成测试覆盖率报告
	@echo "正在生成测试覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告已生成: coverage.html"

# 基准测试
bench: ## 运行基准测试
	@echo "正在运行基准测试..."
	go test -bench=. -benchmem ./...

# 清理
clean: ## 清理编译产物和日志
	@echo "正在清理..."
	rm -rf $(BUILD_DIR)/
	rm -rf logs/*.log
	rm -f coverage.out coverage.html
	@echo "✓ 清理完成"

# 代码检查
lint: ## 运行代码检查
	@echo "正在运行代码检查..."
	golangci-lint run

# 格式化代码
fmt: ## 格式化代码
	@echo "正在格式化代码..."
	go fmt ./...
	@echo "✓ 代码格式化完成"

# 生成 Swagger 文档
swagger: ## 生成 Swagger 文档
	@echo "正在生成 Swagger 文档..."
	swag init -g $(MAIN_FILE) -o ./docs
	@echo "✓ Swagger 文档已生成"

# 安装依赖
deps: ## 安装项目依赖
	@echo "正在安装依赖..."
	go mod download
	go mod tidy
	@echo "✓ 依赖安装完成"

# Docker 构建
docker-build: ## 构建 Docker 镜像
	@echo "正在构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest .
	@echo "✓ Docker 镜像构建完成"

# Docker 运行
docker-run: ## 运行 Docker 容器
	@echo "正在启动 Docker 容器..."
	docker run -d -p 8000:8000 --name $(APP_NAME) $(APP_NAME):latest
	@echo "✓ Docker 容器已启动"

# Docker 停止
docker-stop: ## 停止 Docker 容器
	@echo "正在停止 Docker 容器..."
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)
	@echo "✓ Docker 容器已停止"

# 生产环境编译（优化）
build-prod: ## 编译生产环境版本（优化）
	@echo "正在编译生产环境版本..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✓ 生产环境编译完成"

# Windows 编译
build-windows: ## 编译 Windows 版本
	@echo "正在编译 Windows 版本..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME).exe $(MAIN_FILE)
	@echo "✓ Windows 版本编译完成"

# 帮助信息
help: ## 显示帮助信息
	@echo "Youlai-Gin Makefile 命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
