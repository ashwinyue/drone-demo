# Demo Web App Makefile

# 变量定义
APP_NAME := demo-web-app
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Docker 相关
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_LATEST := $(APP_NAME):latest
DOCKER_REGISTRY := docker.io

# 构建标志
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.GitCommit=$(GIT_COMMIT) \
           -X main.GoVersion=$(GO_VERSION)

# 默认目标
.PHONY: all
all: clean deps test build

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 代码检查
.PHONY: lint
lint:
	@echo "Running linters..."
	go vet ./...
	go fmt ./...

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 构建应用
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o bin/$(APP_NAME) .

# 本地构建（当前平台）
.PHONY: build-local
build-local:
	@echo "Building $(APP_NAME) for local platform..."
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) .

# 运行应用
.PHONY: run
run: build-local
	@echo "Starting $(APP_NAME)..."
	./bin/$(APP_NAME)

# 开发模式运行
.PHONY: dev
dev:
	@echo "Starting $(APP_NAME) in development mode..."
	go run -ldflags "$(LDFLAGS)" . 

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) -t $(DOCKER_LATEST) .

# Docker 运行
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 $(DOCKER_LATEST)

# Docker 推送
.PHONY: docker-push
docker-push: docker-build
	@echo "Pushing Docker image to registry..."
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	docker tag $(DOCKER_LATEST) $(DOCKER_REGISTRY)/$(DOCKER_LATEST)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_LATEST)

# Kubernetes 部署
.PHONY: k8s-deploy
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/

# Kubernetes 删除
.PHONY: k8s-delete
k8s-delete:
	@echo "Deleting from Kubernetes..."
	kubectl delete -f k8s/

# Kubernetes 状态检查
.PHONY: k8s-status
k8s-status:
	@echo "Checking Kubernetes deployment status..."
	kubectl get pods -l app=$(APP_NAME)
	kubectl get services -l app=$(APP_NAME)

# 使用 go-drone-deploy 进行部署
.PHONY: deploy-docker
deploy-docker:
	@echo "Deploying with go-drone-deploy (Docker flow)..."
	../go-drone-deploy/bin/go-drone-deploy -flow=docker -env=dev

.PHONY: deploy-k8s
deploy-k8s:
	@echo "Deploying with go-drone-deploy (K8s flow)..."
	../go-drone-deploy/bin/go-drone-deploy -flow=k8s -env=dev

.PHONY: deploy-full
deploy-full:
	@echo "Deploying with go-drone-deploy (Full flow)..."
	../go-drone-deploy/bin/go-drone-deploy -flow=full -env=dev

.PHONY: deploy-notify
deploy-notify:
	@echo "Sending notification with go-drone-deploy..."
	../go-drone-deploy/bin/go-drone-deploy -flow=notify -env=dev

# 清理
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker rmi $(DOCKER_IMAGE) $(DOCKER_LATEST) 2>/dev/null || true

# 显示版本信息
.PHONY: version
version:
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

# 健康检查
.PHONY: health
health:
	@echo "Checking application health..."
	curl -f http://localhost:8080/health || echo "Application is not running"

# 完整的 CI/CD 流程
.PHONY: ci
ci: clean deps lint test build docker-build

# 发布流程
.PHONY: release
release: ci docker-push
	@echo "Release $(VERSION) completed!"

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Run clean, deps, test, build"
	@echo "  deps         - Install dependencies"
	@echo "  lint         - Run code linters"
	@echo "  test         - Run tests with coverage"
	@echo "  build        - Build application for Linux"
	@echo "  build-local  - Build application for current platform"
	@echo "  run          - Build and run application locally"
	@echo "  dev          - Run application in development mode"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Build and run Docker container"
	@echo "  docker-push  - Push Docker image to registry"
	@echo "  k8s-deploy   - Deploy to Kubernetes"
	@echo "  k8s-delete   - Delete from Kubernetes"
	@echo "  k8s-status   - Check Kubernetes deployment status"
	@echo "  deploy-*     - Deploy using go-drone-deploy tool"
	@echo "  clean        - Clean build artifacts"
	@echo "  version      - Show version information"
	@echo "  health       - Check application health"
	@echo "  ci           - Run CI pipeline"
	@echo "  release      - Run release pipeline"
	@echo "  help         - Show this help message"