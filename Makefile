# Go项目Makefile

# 应用名称
APP_NAME := ad-platform-backend

# Go相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_NAME := $(APP_NAME)
BINARY_UNIX := $(BINARY_NAME)_unix

# 主要目标
.PHONY: all build clean test coverage deps run dev help

all: deps build

# 构建
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o bin/$(BINARY_NAME) -v ./cmd

# 清理
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)

# 测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 测试覆盖率
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 安装依赖
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 运行
run: build
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

# 开发模式运行
dev:
	@echo "Running in development mode..."
	$(GOCMD) run ./cmd

# Linux构建
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v ./cmd

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 代码检查
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# 更新依赖
update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# 创建数据库迁移
migration:
	@echo "Running database migration..."
	$(GOCMD) run ./cmd -migrate

# 生成API文档
docs:
	@echo "Generating API documentation..."
	swag init -g ./cmd/main.go -o ./docs

# 帮助
help:
	@echo "Available commands:"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage"
	@echo "  deps       - Install dependencies"
	@echo "  run        - Build and run the application"
	@echo "  dev        - Run in development mode"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  update     - Update dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  help       - Show this help message"