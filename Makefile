# Mock Object Storage Service Makefile

# 项目信息
PROJECT_NAME := mock-storage
BUILD_DIR := ./bin
CMD_DIR := ./cmd/server

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 构建目标
BINARY_NAME := $(PROJECT_NAME)
BINARY_UNIX := $(BINARY_NAME)_unix
BINARY_WINDOWS := $(BINARY_NAME).exe

.PHONY: all build clean test deps help run

# 默认目标
all: clean deps build

# 构建项目
build:
	@echo "构建项目..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建多平台版本
build-all: build-linux build-windows build-darwin

build-linux:
	@echo "构建Linux版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(CMD_DIR)

build-windows:
	@echo "构建Windows版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_WINDOWS) $(CMD_DIR)

build-darwin:
	@echo "构建MacOS版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin $(CMD_DIR)

# 运行项目
run:
	@echo "运行服务..."
	$(GOCMD) run $(CMD_DIR)

# 运行测试
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

# 安装依赖
deps:
	@echo "安装依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 清理构建文件
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# 代码格式化
fmt:
	@echo "格式化代码..."
	$(GOCMD) fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 生成API文档
docs:
	@echo "生成API文档..."
	swag init -g $(CMD_DIR)/main.go

# 开发模式（热重载）
dev:
	@echo "开发模式..."
	air -c .air.toml

# 安装项目到GOPATH
install: build
	@echo "安装到系统..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# 显示帮助信息
help:
	@echo "Mock Object Storage Service"
	@echo ""
	@echo "可用命令:"
	@echo "  build      构建项目"
	@echo "  build-all  构建所有平台版本"
	@echo "  run        运行服务"
	@echo "  test       运行测试"
	@echo "  deps       安装依赖"
	@echo "  clean      清理构建文件"
	@echo "  fmt        格式化代码"
	@echo "  lint       代码检查"
	@echo "  docs       生成API文档"
	@echo "  dev        开发模式"
	@echo "  install    安装到系统"
	@echo "  help       显示帮助信息"