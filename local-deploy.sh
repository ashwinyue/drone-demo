#!/bin/bash

# 本地 Drone + Kind 部署脚本
# 使用方法: ./local-deploy.sh

set -e

echo "🚀 启动本地 Drone + Kind 部署环境"
echo "===================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}📋 检查依赖...${NC}"
    
    local missing_deps=()
    
    if ! command -v kind &> /dev/null; then
        missing_deps+=("kind")
    fi
    
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${RED}❌ 缺少以下依赖: ${missing_deps[*]}${NC}"
        echo -e "${YELLOW}请先安装缺少的依赖:${NC}"
        echo "brew install kind kubectl docker go"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 所有依赖已安装${NC}"
}

# 创建 Kind 配置文件
create_kind_config() {
    echo -e "${BLUE}📄 创建 Kind 配置文件...${NC}"
    
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
EOF

    echo -e "${GREEN}✅ Kind 配置文件已创建${NC}"
}

# 创建 Kind 集群
setup_kind_cluster() {
    echo -e "${BLUE}🔧 设置 Kind 集群...${NC}"
    
    if kind get clusters 2>/dev/null | grep -q "drone-demo"; then
        echo -e "${YELLOW}📦 Kind 集群已存在，跳过创建${NC}"
    else
        echo -e "${BLUE}📦 创建 Kind 集群...${NC}"
        kind create cluster --config=kind-config.yaml
        
        # 等待集群就绪
        echo -e "${BLUE}⏳ 等待集群就绪...${NC}"
        kubectl wait --for=condition=Ready nodes --all --timeout=300s
    fi
    
    echo -e "${GREEN}✅ Kind 集群就绪${NC}"
}

# 设置本地 Registry
setup_local_registry() {
    echo -e "${BLUE}🐳 设置本地 Docker Registry...${NC}"
    
    if docker ps --format "table {{.Names}}" | grep -q "^registry$"; then
        echo -e "${YELLOW}📦 Registry 已运行，跳过创建${NC}"
    else
        echo -e "${BLUE}📦 启动本地 Registry...${NC}"
        docker run -d --restart=always -p 5000:5000 --name registry registry:2
        
        # 连接到 kind 网络
        docker network connect "kind" registry 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✅ 本地 Registry 就绪${NC}"
}

# 创建本地 Docker Compose 配置
create_drone_config() {
    echo -e "${BLUE}📄 创建 Drone 配置文件...${NC}"
    
    # 生成随机 RPC Secret
    local RPC_SECRET=$(openssl rand -hex 16)
    local GITHUB_USERNAME=$(git config user.name 2>/dev/null || echo "admin")
    
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
      - DRONE_GITHUB_CLIENT_ID=\${GITHUB_CLIENT_ID:-dummy}
      - DRONE_GITHUB_CLIENT_SECRET=\${GITHUB_CLIENT_SECRET:-dummy}
      - DRONE_RPC_SECRET=${RPC_SECRET}
      - DRONE_SERVER_HOST=localhost:8080
      - DRONE_SERVER_PROTO=http
      - DRONE_USER_CREATE=username:${GITHUB_USERNAME},admin:true
      - DRONE_LOGS_DEBUG=true
      - DRONE_OPEN=true
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
      - DRONE_RPC_SECRET=${RPC_SECRET}
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

    echo -e "${GREEN}✅ Drone 配置文件已创建${NC}"
    echo -e "${YELLOW}🔑 RPC Secret: ${RPC_SECRET}${NC}"
}

# 启动 Drone 服务
start_drone() {
    echo -e "${BLUE}🚁 启动 Drone 服务...${NC}"
    
    docker-compose -f docker-compose-local.yml up -d
    
    # 等待服务启动
    echo -e "${BLUE}⏳ 等待 Drone 服务启动...${NC}"
    sleep 10
    
    # 检查服务状态
    if curl -s http://localhost:8080 > /dev/null; then
        echo -e "${GREEN}✅ Drone 服务已启动${NC}"
        echo -e "${BLUE}📱 访问地址: http://localhost:8080${NC}"
    else
        echo -e "${YELLOW}⚠️  Drone 服务可能还在启动中，请稍后访问${NC}"
    fi
}

# 构建并部署应用
deploy_app() {
    echo -e "${BLUE}🚀 构建并部署应用...${NC}"
    
    # 进入应用目录
    cd demo-web-app
    
    # 构建 Go 应用
    echo -e "${BLUE}🔨 构建 Go 应用...${NC}"
    CGO_ENABLED=0 GOOS=linux go build -o demo-web-app .
    
    # 构建 Docker 镜像
    echo -e "${BLUE}🐳 构建 Docker 镜像...${NC}"
    docker build -t localhost:5000/demo-web-app:latest .
    
    # 推送到本地 Registry
    echo -e "${BLUE}📤 推送镜像到本地 Registry...${NC}"
    docker push localhost:5000/demo-web-app:latest
    
    # 部署到 Kubernetes
    echo -e "${BLUE}☸️  部署到 Kubernetes...${NC}"
    kubectl apply -f k8s/
    
    # 更新镜像
    kubectl set image deployment/demo-web-app demo-web-app=localhost:5000/demo-web-app:latest
    
    # 等待部署完成
    echo -e "${BLUE}⏳ 等待部署完成...${NC}"
    kubectl rollout status deployment/demo-web-app --timeout=300s
    
    cd ..
    
    echo -e "${GREEN}✅ 应用部署完成${NC}"
}

# 创建本地 .drone.yml
create_local_drone_yml() {
    echo -e "${BLUE}📄 创建本地 Drone 配置...${NC}"
    
    cat > .drone-local.yml << 'EOF'
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
      path: ~/.kube/config

trigger:
  event:
    - push
    - pull_request
EOF

    echo -e "${GREEN}✅ 本地 Drone 配置已创建${NC}"
}

# 显示环境状态
show_status() {
    echo -e "${BLUE}📊 环境状态:${NC}"
    echo ""
    
    echo -e "${YELLOW}Kind 集群:${NC}"
    kubectl get nodes
    echo ""
    
    echo -e "${YELLOW}应用状态:${NC}"
    kubectl get pods,svc
    echo ""
    
    echo -e "${YELLOW}Drone 服务:${NC}"
    docker-compose -f docker-compose-local.yml ps
    echo ""
}

# 显示访问信息
show_access_info() {
    echo -e "${GREEN}🎉 本地部署环境设置完成！${NC}"
    echo ""
    echo -e "${BLUE}📱 访问地址:${NC}"
    echo "  - Drone 界面: http://localhost:8080"
    echo "  - 应用访问: http://localhost:30000"
    echo "  - 健康检查: http://localhost:30000/health"
    echo "  - 应用信息: http://localhost:30000/api/info"
    echo ""
    echo -e "${YELLOW}📋 后续步骤:${NC}"
    echo "  1. 访问 Drone 界面并激活仓库"
    echo "  2. 修改代码并推送以触发自动部署"
    echo "  3. 查看 KIND_LOCAL_SETUP.md 了解更多详情"
    echo ""
    echo -e "${YELLOW}🧪 测试命令:${NC}"
    echo "  curl http://localhost:30000"
    echo "  curl http://localhost:30000/health"
    echo ""
    echo -e "${YELLOW}🧹 清理命令:${NC}"
    echo "  ./local-deploy.sh cleanup"
}

# 清理环境
cleanup() {
    echo -e "${BLUE}🧹 清理本地环境...${NC}"
    
    # 停止 Drone 服务
    if [ -f "docker-compose-local.yml" ]; then
        docker-compose -f docker-compose-local.yml down -v
    fi
    
    # 删除 Kind 集群
    if kind get clusters 2>/dev/null | grep -q "drone-demo"; then
        kind delete cluster --name drone-demo
    fi
    
    # 停止本地 Registry
    if docker ps --format "table {{.Names}}" | grep -q "^registry$"; then
        docker stop registry
        docker rm registry
    fi
    
    # 清理配置文件
    rm -f kind-config.yaml docker-compose-local.yml .drone-local.yml
    
    echo -e "${GREEN}✅ 环境清理完成${NC}"
}

# 主函数
main() {
    case "${1:-setup}" in
        "cleanup")
            cleanup
            ;;
        "setup"|"")
            check_dependencies
            create_kind_config
            setup_kind_cluster
            setup_local_registry
            create_drone_config
            start_drone
            create_local_drone_yml
            deploy_app
            show_status
            show_access_info
            ;;
        *)
            echo "使用方法: $0 [setup|cleanup]"
            echo "  setup   - 设置本地环境（默认）"
            echo "  cleanup - 清理本地环境"
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"