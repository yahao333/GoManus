# GoManus Makefile
# 用于构建、测试和部署 GoManus 项目

# 变量定义
BINARY_NAME=gomanus
MAIN_PATH=./cmd
BUILD_DIR=build
VERSION?=0.1.0
BUILD_TIME=$(shell date +%Y-%m-%d)
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY=$(shell git diff --shortstat 2>/dev/null || true)
GO_FILES=$(shell find . -name "*.go" -type f | grep -v vendor/)
CONFIG_DIR=config
WORKSPACE_DIR=workspace
LOGS_DIR=logs

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go 设置
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=1

# 默认目标
.PHONY: all
all: clean deps build

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "==================================="
	@echo "    🤖 GoManus - AI Agent 框架"
	@echo "==================================="
	@echo ""
	@echo "🎯 快速开始:"
	@echo "  make dev-setup    # 设置开发环境"
	@echo "  make build        # 构建项目"
	@echo "  make run          # 运行项目"
	@echo ""
	@echo "🔧 开发命令:"
	@echo "  make init-config  # 初始化配置文件"
	@echo "  make run-agent    # 运行智能体"
	@echo "  make example      # 运行示例"
	@echo ""
	@echo "📊 质量检查:"
	@echo "  make check        # 运行所有检查"
	@echo "  make test         # 运行测试"
	@echo "  make fmt          # 格式化代码"
	@echo ""
	@echo "📦 构建发布:"
	@echo "  make build-all    # 构建所有平台"
	@echo "  make release      # 创建发布包"
	@echo ""
	@echo "所有可用命令:"
	@echo "-------------------"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "💡 提示: 使用 'make <命令>' 来执行相应操作"

# 依赖管理
.PHONY: deps
deps: ## 安装 Go 依赖
	@echo "📦 安装依赖..."
	go mod download
	go mod verify
	go mod tidy

# 构建
.PHONY: build
build: ## 构建应用程序
	@echo "🔨 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "📋 运行方式:"
	@echo "  $(BUILD_DIR)/$(BINARY_NAME) --help    # 查看帮助"
	@echo "  $(BUILD_DIR)/$(BINARY_NAME) run       # 运行智能体"
	@echo "  $(BUILD_DIR)/$(BINARY_NAME) config    # 配置管理"

# 构建特定平台
.PHONY: build-linux
build-linux: ## 构建 Linux 版本
	@echo "🐧 构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Linux 版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-darwin
build-darwin: ## 构建 macOS 版本
	@echo "🍎 构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "✅ macOS 版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-darwin-*"

.PHONY: build-windows
build-windows: ## 构建 Windows 版本
	@echo "🪟 构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Windows 版本构建完成: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# 构建所有平台
.PHONY: build-all
build-all: ## 构建所有平台版本
	@echo "🌍 构建所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ 所有平台构建完成"
	@ls -la $(BUILD_DIR)/

# 运行
.PHONY: run
run: ## 运行应用程序（开发模式）
	@echo "🚀 运行 $(BINARY_NAME)..."
	go run $(MAIN_PATH)

.PHONY: run-help
run-help: ## 运行并显示帮助
	@echo "📖 运行并显示帮助信息..."
	go run $(MAIN_PATH) --help

.PHONY: run-version
run-version: ## 运行并显示版本
	@echo "🔖 运行并显示版本信息..."
	go run $(MAIN_PATH) --version

# 测试
.PHONY: test
test: ## 运行测试
	@echo "🧪 运行测试..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "📊 运行测试覆盖率分析..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📈 覆盖率报告已生成: coverage.html"

# 代码质量
.PHONY: fmt
fmt: ## 格式化 Go 代码
	@echo "🎨 格式化代码..."
	go fmt ./...
	@echo "✅ 代码格式化完成"

.PHONY: vet
vet: ## 运行 go vet 检查
	@echo "🔍 运行代码检查..."
	go vet ./...
	@echo "✅ 代码检查完成"

.PHONY: lint
lint: ## 运行代码质量检查
	@echo "🧹 运行代码质量检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过代码质量检查"; \
		echo "💡 安装: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2"; \
	fi

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "🗑️  清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.test *.prof
	@echo "✅ 清理完成"

# 深度清理
.PHONY: clean-all
clean-all: clean ## 清理所有生成的文件
	@echo "🗑️  深度清理..."
	@rm -rf vendor/
	@go clean -cache -modcache -testcache
	@echo "✅ 深度清理完成"

# 安装
.PHONY: install
install: build ## 安装到 GOPATH/bin
	@echo "📦 安装 $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PATH)
	@echo "✅ 安装完成"

# 开发环境
.PHONY: dev-setup
dev-setup: ## 设置开发环境
	@echo "🔧 设置开发环境..."
	@echo "1. 安装依赖..."
	go mod download
	@echo "2. 安装开发工具..."
	@which golangci-lint >/dev/null || (echo "安装 golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin)
	@echo "3. 创建必要的目录..."
	@mkdir -p $(WORKSPACE_DIR) $(LOGS_DIR)
	@echo "4. 初始化配置文件..."
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		go run $(MAIN_PATH) config init; \
	fi
	@echo "✅ 开发环境设置完成"
	@echo "📋 快速开始:"
	@echo "  make run-help     # 查看命令帮助"
	@echo "  make run          # 运行交互模式"
	@echo "  make build        # 构建项目"

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "📋 GoManus 版本信息:"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git 提交: $(GIT_COMMIT)"
	@if [ -n "$(GIT_DIRTY)" ]; then \
		echo "  状态: 包含未提交的更改"; \
	else \
		echo "  状态: 干净"; \
	fi

# 发布构建
.PHONY: release
release: clean test build-all ## 构建发布版本
	@echo "📦 准备发布包..."
	@mkdir -p $(BUILD_DIR)/release
	# 创建压缩包
	@cd $(BUILD_DIR) && \
		tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip -q release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "✅ 发布包已创建:"
	@ls -la $(BUILD_DIR)/release/

# Docker
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "🐳 构建 Docker 镜像..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "✅ Docker 镜像构建完成"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "🐳 运行 Docker 容器..."
	docker run --rm -it \
		-v $$(pwd)/config:/app/config \
		-v $$(pwd)/logs:/app/logs \
		-v $$(pwd)/workspace:/app/workspace \
		$(BINARY_NAME):latest

# 文档生成
.PHONY: docs
docs: ## 生成文档
	@echo "📚 生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "启动 godoc 服务器..."; \
		echo "访问: http://localhost:6060/pkg/github.com/yahao333/GoManus/"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc 未安装"; \
		echo "安装: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 示例运行
.PHONY: example
example: ## 运行示例程序
	@echo "🎯 运行示例程序..."
	go run examples/main.go

.PHONY: example-build
example-build: ## 构建示例程序
	@echo "🔨 构建示例程序..."
	go build -o $(BUILD_DIR)/gomanus-example examples/main.go
	@echo "✅ 示例程序构建完成: $(BUILD_DIR)/gomanus-example"

# 性能分析
.PHONY: benchmark
benchmark: ## 运行性能测试
	@echo "📊 运行性能测试..."
	go test -bench=. -benchmem ./...

# 依赖更新
.PHONY: update-deps
update-deps: ## 更新依赖
	@echo "🔄 更新依赖..."
	go get -u ./...
	go mod tidy
	@echo "✅ 依赖更新完成"

# 安全检查
.PHONY: security
security: ## 运行安全检查
	@echo "🔒 运行安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec 未安装，跳过安全检查"; \
		echo "💡 安装: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 项目信息
.PHONY: info
info: ## 显示项目信息
	@echo "📋 GoManus 项目信息:"
	@echo "  名称: $(BINARY_NAME)"
	@echo "  Go 版本: $(shell go version)"
	@echo "  GOOS: $(GOOS)"
	@echo "  GOARCH: $(GOARCH)"
	@echo "  GOPATH: $(shell go env GOPATH)"
	@echo "  GOCACHE: $(shell go env GOCACHE)"
	@echo "  GOMODCACHE: $(shell go env GOMODCACHE)"

# 快速检查
.PHONY: check
check: fmt vet test security ## 运行所有检查（格式化、代码检查、测试、安全）

# 快速构建和运行
.PHONY: quick
quick: clean build run-help ## 快速构建并显示帮助信息

# 完整构建流程
.PHONY: full-build
full-build: clean deps check build-all ## 完整构建流程（清理、依赖、检查、构建）

# 监控（开发时使用）
.PHONY: watch
watch: ## 监控文件变化并重新运行
	@echo "👀 监控文件变化..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air 未安装，跳过监控"; \
		echo "💡 安装: go install github.com/cosmtrek/air@latest"; \
	fi

# GoManus 特定功能
.PHONY: init-config
init-config: ## 初始化配置文件
	@echo "⚙️  初始化配置文件..."
	go run $(MAIN_PATH) config init

.PHONY: validate-config
validate-config: ## 验证配置文件
	@echo "✅ 验证配置文件..."
	go run $(MAIN_PATH) config validate

.PHONY: run-agent
run-agent: ## 运行智能体（交互模式）
	@echo "🤖 运行智能体..."
	go run $(MAIN_PATH) run

.PHONY: run-direct
run-direct: ## 直接运行模式
	@echo "🚀 直接运行模式..."
	go run $(MAIN_PATH) direct

# 默认目标
.DEFAULT_GOAL := help
