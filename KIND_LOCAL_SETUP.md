# Kind 本地 Drone 自动部署指南

本指南将帮助您在本地使用 Kind（Kubernetes in Docker）集群实现 Drone CI/CD 自动部署。

## 🎯 架构概述

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitHub Repo   │───▶│   Drone CI/CD   │───▶│   Kind Cluster  │
│                 │    │                 │    │                 │
│ - demo-web-app  │    │ - Build & Test  │    │ - Local K8s     │
│ - .drone.yml    │    │ - Docker Build  │    │ - Pod & Service │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 前置条件

- ✅ Docker Desktop 已安装并运行
- ✅ Kind 已安装
- ✅ kubectl 已安装
- ✅ Go 1.21+ 已安装

## 🚀 快速开始

### 方式一：一键部署（推荐）

我们提供了一键部署脚本，可以自动设置完整的本地环境：

```bash
# 运行一键部署脚本
./local-deploy.sh

# 清理环境
./local-deploy.sh cleanup
```

脚本会自动完成以下操作：
- 检查并安装必要依赖
- 创建 Kind 集群
- 设置本地 Docker Registry
- 启动 Drone 服务
- 构建并部署应用
- 显示访问信息

### 方式二：手动安装

#### 1. 安装 Kind

```bash
# macOS 使用 Homebrew
brew install kind

# 或者直接下载二进制文件
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### 2. 创建 Kind 集群

```bash
# 创建集群配置文件
cat > kind-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: drone-demo
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
  - containerPort: 30000
    hostPort: 30000
    protocol: TCP
- role: worker
- role: worker
EOF

# 创建集群
kind create cluster --config=kind-config.yaml

# 验证集群
kubectl cluster-info --context kind-drone-demo
kubectl get nodes
```

### 3. 设置本地 Docker Registry

```bash
# 启动本地 Docker Registry
docker run -d --restart=always -p 5000:5000 --name registry registry:2

# 连接 Registry 到 Kind 网络
docker network connect "kind" registry
```

### 4. 配置 Drone 服务

创建本地 Drone 配置：

```bash
# 创建 docker-compose-local.yml
cat > docker-compose-local.yml << EOF
version: '3.8'

services:
  drone-server:
    image: drone/drone:2
    container_name: drone-server-local
    ports:
      - "8080:80"
    volumes:
      - drone-data-local:/data
    environment:
      - DRONE_GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - DRONE_GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - DRONE_RPC_SECRET=${RPC_SECRET:-$(openssl rand -hex 16)}
      - DRONE_SERVER_HOST=localhost:8080
      - DRONE_SERVER_PROTO=http
      - DRONE_USER_CREATE=username:${GITHUB_USERNAME:-admin},admin:true
      - DRONE_LOGS_DEBUG=true
    restart: unless-stopped
    networks:
      - kind

  drone-runner:
    image: drone/drone-runner-docker:1
    container_name: drone-runner-local
    ports:
      - "3000:3000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_RPC_PROTO=http
      - DRONE_RPC_HOST=drone-server-local
      - DRONE_RPC_SECRET=${RPC_SECRET:-$(openssl rand -hex 16)}
      - DRONE_RUNNER_CAPACITY=2
      - DRONE_RUNNER_NAME=local-runner
    restart: unless-stopped
    depends_on:
      - drone-server
    networks:
      - kind

volumes:
  drone-data-local:

networks:
  kind:
    external: true
EOF
```

### 5. 修改部署配置

更新 `demo-web-app/deploy-config.yaml` 以适配本地环境：

```yaml
server:
  http:
    addr: 0.0.0.0:8080
    timeout: 1s

deploy:
  project_name: "demo-web-app"
  author: "local-dev"
  namespace: "default"
  version: "latest"
  env: "local"
  
  docker:
    registry: "localhost:5000"
    username: ""
    password: ""
    image_name: "demo-web-app:latest"
    dockerfile_path: "./Dockerfile"
    build_context: "."
  
  k8s:
    kubeconfig_path: "~/.kube/config"
    namespace: "default"
    deployment_name: "demo-web-app"
    service_name: "demo-web-app-service"
    replicas: 1
    resources:
      requests:
        memory: "32Mi"
        cpu: "25m"
      limits:
        memory: "64Mi"
        cpu: "50m"
    ports:
      - name: "http"
        port: 80
        target_port: 8080
        protocol: "TCP"
        node_port: 30000
    env_vars:
      - name: "APP_ENV"
        value: "local"
      - name: "PORT"
        value: "8080"
  
  notify:
    enabled: false
```

### 6. 创建本地 .drone.yml

为本地环境创建专门的 Drone 配置：

```yaml
# .drone-local.yml
kind: pipeline
type: docker
name: local-build-deploy

steps:
  # 测试阶段
  - name: test
    image: golang:1.21-alpine
    commands:
      - cd demo-web-app
      - go mod download
      - go test ./...
      - go vet ./...
    when:
      event: [push, pull_request]

  # 构建阶段
  - name: build
    image: golang:1.21-alpine
    commands:
      - cd demo-web-app
      - go mod download
      - CGO_ENABLED=0 GOOS=linux go build -o demo-web-app .
    when:
      event: [push, pull_request]

  # 本地 Docker 构建和推送
  - name: docker-build-local
    image: docker:dind
    volumes:
      - name: docker-sock
        path: /var/run/docker.sock
    commands:
      - cd demo-web-app
      - docker build -t localhost:5000/demo-web-app:latest .
      - docker push localhost:5000/demo-web-app:latest
    when:
      event: [push]
      branch: [main, develop]

  # Kind 集群部署
  - name: k8s-deploy-local
    image: bitnami/kubectl:latest
    environment:
      KUBECONFIG: /tmp/kubeconfig
    volumes:
      - name: kubeconfig
        path: /tmp/kubeconfig
    commands:
      - kubectl apply -f demo-web-app/k8s/
      - kubectl set image deployment/demo-web-app demo-web-app=localhost:5000/demo-web-app:latest
      - kubectl rollout status deployment/demo-web-app
      - kubectl get pods
      - kubectl get services
    when:
      event: [push]
      branch: [main, develop]

volumes:
  - name: docker-sock
    host:
      path: /var/run/docker.sock
  - name: kubeconfig
    host:
      path: /Users/${USER}/.kube/config

trigger:
  event:
    - push
    - pull_request
```

## 🔧 本地部署脚本

创建自动化部署脚本：

```bash
#!/bin/bash
# local-deploy.sh

set -e

echo "🚀 启动本地 Drone + Kind 部署环境"

# 检查依赖
check_dependencies() {
    echo "📋 检查依赖..."
    
    if ! command -v kind &> /dev/null; then
        echo "❌ Kind 未安装"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        echo "❌ kubectl 未安装"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker 未安装"
        exit 1
    fi
    
    echo "✅ 所有依赖已安装"
}

# 创建 Kind 集群
setup_kind_cluster() {
    echo "🔧 设置 Kind 集群..."
    
    if kind get clusters | grep -q "drone-demo"; then
        echo "📦 Kind 集群已存在"
    else
        echo "📦 创建 Kind 集群..."
        kind create cluster --config=kind-config.yaml
    fi
    
    # 等待集群就绪
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    echo "✅ Kind 集群就绪"
}

# 设置本地 Registry
setup_local_registry() {
    echo "🐳 设置本地 Docker Registry..."
    
    if docker ps | grep -q "registry"; then
        echo "📦 Registry 已运行"
    else
        docker run -d --restart=always -p 5000:5000 --name registry registry:2
        docker network connect "kind" registry
    fi
    
    echo "✅ 本地 Registry 就绪"
}

# 启动 Drone 服务
start_drone() {
    echo "🚁 启动 Drone 服务..."
    
    export RPC_SECRET=$(openssl rand -hex 16)
    export GITHUB_USERNAME=$(git config user.name)
    
    docker-compose -f docker-compose-local.yml up -d
    
    echo "✅ Drone 服务已启动"
    echo "📱 访问地址: http://localhost:8080"
}

# 部署应用到 Kind
deploy_to_kind() {
    echo "🚀 部署应用到 Kind..."
    
    # 构建并推送镜像
    cd demo-web-app
    docker build -t localhost:5000/demo-web-app:latest .
    docker push localhost:5000/demo-web-app:latest
    
    # 部署到 Kubernetes
    kubectl apply -f k8s/
    kubectl set image deployment/demo-web-app demo-web-app=localhost:5000/demo-web-app:latest
    kubectl rollout status deployment/demo-web-app
    
    echo "✅ 应用部署完成"
    echo "🌐 访问地址: http://localhost:30000"
}

# 显示状态
show_status() {
    echo "📊 环境状态:"
    echo "Kind 集群:"
    kubectl get nodes
    echo ""
    echo "应用状态:"
    kubectl get pods,svc
    echo ""
    echo "Drone 服务:"
    docker-compose -f docker-compose-local.yml ps
}

# 主函数
main() {
    check_dependencies
    setup_kind_cluster
    setup_local_registry
    start_drone
    deploy_to_kind
    show_status
    
    echo "🎉 本地部署环境设置完成！"
    echo "📱 Drone 界面: http://localhost:8080"
    echo "🌐 应用访问: http://localhost:30000"
}

# 运行主函数
main
```

## 🧪 测试本地部署

### 1. 运行部署脚本

```bash
# 给脚本执行权限
chmod +x local-deploy.sh

# 运行部署
./local-deploy.sh
```

### 2. 验证部署

```bash
# 检查 Kind 集群
kubectl get nodes
kubectl get pods -A

# 检查应用
kubectl get pods,svc
kubectl logs -f deployment/demo-web-app

# 测试应用访问
curl http://localhost:30000
curl http://localhost:30000/health
```

### 3. 测试 Drone 流水线

```bash
# 修改代码触发构建
echo "// Local test $(date)" >> demo-web-app/main.go

# 提交推送
git add .
git commit -m "Test local drone deployment"
git push origin main
```

## 🔍 监控和调试

### 查看日志

```bash
# Drone 服务日志
docker logs drone-server-local
docker logs drone-runner-local

# Kind 集群日志
kubectl logs -f deployment/demo-web-app

# Docker Registry 日志
docker logs registry
```

### 调试命令

```bash
# 进入 Kind 节点
docker exec -it drone-demo-control-plane bash

# 查看镜像
docker exec -it drone-demo-control-plane crictl images

# 手动部署测试
kubectl apply -f demo-web-app/k8s/
kubectl get events --sort-by=.metadata.creationTimestamp
```

## 🧹 清理环境

```bash
# 停止 Drone 服务
docker-compose -f docker-compose-local.yml down

# 删除 Kind 集群
kind delete cluster --name drone-demo

# 停止本地 Registry
docker stop registry
docker rm registry

# 清理 Docker 镜像
docker system prune -f
```

## 🎯 本地开发工作流

1. **开发代码** → 修改 `demo-web-app/main.go`
2. **本地测试** → `go run demo-web-app/main.go`
3. **提交代码** → `git commit && git push`
4. **自动构建** → Drone 自动触发流水线
5. **自动部署** → 部署到本地 Kind 集群
6. **验证结果** → 访问 `http://localhost:30000`

## 📚 相关资源

- [Kind 官方文档](https://kind.sigs.k8s.io/)
- [Drone 本地开发指南](https://docs.drone.io/)
- [Kubernetes 本地开发最佳实践](https://kubernetes.io/docs/setup/learning-environment/)

---

**🎉 现在您可以在本地完整体验 Drone + Kubernetes 的 CI/CD 流程了！**