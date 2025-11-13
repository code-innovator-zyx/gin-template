.PHONY: help run build test clean swagger install lint

# 默认目标
.DEFAULT_GOAL := help

# 应用名称
APP_NAME := gin-template
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)

# Go 相关变量
GO := go
GOFLAGS := -v
LDFLAGS := -w -s

help: ## 显示帮助信息
	@echo "可用的命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

rename: ## 重命名项目 (用法: make rename NEW_NAME=your-project-name)
	@if [ -z "$(NEW_NAME)" ]; then \
		echo "错误: 请提供新的项目名称"; \
		echo "用法: make rename NEW_NAME=your-project-name"; \
		exit 1; \
	fi
	@echo "正在将项目从 'gin-template' 重命名为 '$(NEW_NAME)'..."
	@echo ""
	@echo "步骤 1/5: 更新 go.mod 模块名..."
	@sed -i.bak 's|module gin-template|module $(NEW_NAME)|g' go.mod && rm go.mod.bak
	@echo "✓ go.mod 已更新"
	@echo ""
	@echo "步骤 2/5: 更新所有 Go 文件中的 import 路径..."
	@find . -type f -name "*.go" ! -path "*/vendor/*" ! -path "*/\.*" -exec sed -i.bak 's|gin-template/|$(NEW_NAME)/|g' {} \; -exec rm {}.bak \;
	@echo "✓ Go 文件导入路径已更新"
	@echo ""
	@echo "步骤 3/5: 更新 Makefile..."
	@sed -i.bak 's|APP_NAME := gin-template|APP_NAME := $(NEW_NAME)|g' Makefile && rm Makefile.bak
	@echo "✓ Makefile 已更新"
	@echo ""
	@echo "步骤 4/5: 更新 docker-compose.yml..."
	@if [ -f docker-compose.yml ]; then \
		sed -i.bak 's|gin-template|$(NEW_NAME)|g' docker-compose.yml && rm docker-compose.yml.bak; \
		echo "✓ docker-compose.yml 已更新"; \
	fi
	@echo ""
	@echo "步骤 5/5: 更新文档..."
	@find . -type f \( -name "*.md" \) ! -path "*/vendor/*" ! -path "*/\.*" -exec sed -i.bak 's|gin-template|$(NEW_NAME)|g' {} \; -exec rm {}.bak \;
	@echo "✓ 文档已更新"
	@echo ""
	@echo "=========================================="
	@echo "✅ 重命名完成！"
	@echo "=========================================="
	@echo ""
	@echo "项目已从 'gin-template' 重命名为 '$(NEW_NAME)'"
	@echo ""
	@echo "下一步操作："
	@echo "  1. 运行: go mod tidy"
	@echo "  2. 运行: make init-config (如果还没有 app.yaml)"
	@echo "  3. 运行: make run"
	@echo ""
	@echo "提示: 如果使用 Git，建议执行:"
	@echo "  git add ."
	@echo "  git commit -m 'chore: rename project to $(NEW_NAME)'"
	@echo ""

run: ## 运行应用
	$(GO) run main.go

build: ## 编译应用
	@echo "正在编译..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .
	@echo "编译完成: $(BINARY)"

build-linux: ## 编译Linux版本
	@echo "正在编译Linux版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)-linux-amd64 .
	@echo "编译完成: $(BINARY)-linux-amd64"

build-darwin: ## 编译macOS版本
	@echo "正在编译macOS版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-amd64 .
	@echo "编译完成: $(BINARY)-darwin-amd64"

build-windows: ## 编译Windows版本
	@echo "正在编译Windows版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)-windows-amd64.exe .
	@echo "编译完成: $(BINARY)-windows-amd64.exe"

build-all: build-linux build-darwin build-windows ## 编译所有平台版本

test: ## 运行测试
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## 运行测试并显示覆盖率
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "清理完成"

swagger_v1: ## 生成Swagger文档
	@echo "生成Swagger文档..."
	swag init -g internal/handler/v1/routes.go -o docs --parseDependency --parseInternal --instanceName v1
	@echo "Swagger文档生成完成"

install: ## 安装依赖
	@echo "安装Go依赖..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "依赖安装完成"

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "请先安装 golangci-lint"; exit 1; }
	golangci-lint run ./...

fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GO) fmt ./...
	@echo "格式化完成"

vet: ## 运行go vet
	@echo "运行静态分析..."
	$(GO) vet ./...

dev: ## 开发模式（热重载需要安装air）
	@command -v air >/dev/null 2>&1 || { echo "请先安装 air: go install github.com/cosmtrek/air@latest"; exit 1; }
	air

docker-build: ## 构建Docker镜像
	docker build -t $(APP_NAME):latest .

docker-run: ## 运行Docker容器
	docker run -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest

docker-stop: ## 停止Docker容器
	docker stop $(APP_NAME) && docker rm $(APP_NAME)

up: ## 启动docker-compose服务
	docker-compose up -d

down: ## 停止docker-compose服务
	docker-compose down

logs: ## 查看docker-compose日志
	docker-compose logs -f

init-config: ## 初始化配置文件
	@if [ ! -f app.yaml ]; then \
		echo "创建配置文件..."; \
		cp app.yaml.template app.yaml; \
		echo "配置文件已创建: app.yaml"; \
		echo "请编辑 app.yaml 配置数据库和其他设置"; \
	else \
		echo "配置文件已存在: app.yaml"; \
	fi

check: fmt vet lint ## 运行所有检查

all: clean install check build test ## 执行完整的构建流程
