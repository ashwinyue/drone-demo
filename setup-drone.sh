#!/bin/bash

# Drone CI/CD 快速设置脚本
# 使用方法: ./setup-drone.sh

set -e

echo "🚀 Drone CI/CD 快速设置脚本"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker 未安装，请先安装 Docker${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Docker 已安装${NC}"
}

# 检查 kubectl 是否安装
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        echo -e "${YELLOW}⚠️  kubectl 未安装，Kubernetes 部署功能将不可用${NC}"
    else
        echo -e "${GREEN}✅ kubectl 已安装${NC}"
    fi
}

# 获取用户输入
get_user_input() {
    echo -e "${BLUE}📝 请输入以下配置信息：${NC}"
    
    read -p "Drone 服务器域名 (例: drone.example.com): " DRONE_SERVER_HOST
    read -p "GitHub Client ID: " GITHUB_CLIENT_ID
    read -s -p "GitHub Client Secret: " GITHUB_CLIENT_SECRET
    echo
    read -p "RPC Secret (随机字符串): " RPC_SECRET
    
    if [ -z "$RPC_SECRET" ]; then
        RPC_SECRET=$(openssl rand -hex 16)
        echo -e "${YELLOW}🔑 自动生成 RPC Secret: $RPC_SECRET${NC}"
    fi
}

# 创建 Docker Compose 文件
create_docker_compose() {
    echo -e "${BLUE}📄 创建 Docker Compose 配置文件...${NC}"
    
    cat > docker-compose.yml << EOF
version: '3.8'

services:
  drone-server:
    image: drone/drone:2
    container_name: drone-server
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - drone-data:/data
    environment:
      - DRONE_GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - DRONE_GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - DRONE_RPC_SECRET=${RPC_SECRET}
      - DRONE_SERVER_HOST=${DRONE_SERVER_HOST}
      - DRONE_SERVER_PROTO=https
      - DRONE_USER_CREATE=username:${GITHUB_USERNAME:-admin},admin:true
    restart: unless-stopped

  drone-runner:
    image: drone/drone-runner-docker:1
    container_name: drone-runner
    ports:
      - "3000:3000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_RPC_PROTO=https
      - DRONE_RPC_HOST=${DRONE_SERVER_HOST}
      - DRONE_RPC_SECRET=${RPC_SECRET}
      - DRONE_RUNNER_CAPACITY=2
      - DRONE_RUNNER_NAME=docker-runner
    restart: unless-stopped
    depends_on:
      - drone-server

volumes:
  drone-data:
EOF

    echo -e "${GREEN}✅ Docker Compose 文件已创建${NC}"
}

# 启动 Drone 服务
start_drone() {
    echo -e "${BLUE}🚀 启动 Drone 服务...${NC}"
    
    docker-compose up -d
    
    echo -e "${GREEN}✅ Drone 服务已启动${NC}"
    echo -e "${BLUE}📱 访问地址: https://${DRONE_SERVER_HOST}${NC}"
}

# 创建示例 secrets 配置
create_secrets_example() {
    echo -e "${BLUE}📝 创建 Secrets 配置示例...${NC}"
    
    cat > secrets-example.md << 'EOF'
# Drone Secrets 配置示例

在 Drone 界面中添加以下 Secrets：

## Docker 相关
```
docker_username: your_dockerhub_username
docker_password: your_dockerhub_password_or_token
```

## Kubernetes 相关
```bash
# 获取 kubeconfig 的 base64 编码
kubeconfig_dev: $(cat ~/.kube/config | base64 -w 0)
kubeconfig_prod: $(cat ~/.kube/config-prod | base64 -w 0)
kubeconfig: $(cat ~/.kube/config | base64 -w 0)
```

## 通知相关
```
webhook_url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
github_token: your_github_personal_access_token
```

## 添加 Secrets 的步骤
1. 访问 Drone 界面
2. 进入仓库设置
3. 点击 "Secrets" 选项卡
4. 添加上述 Secrets
EOF

    echo -e "${GREEN}✅ Secrets 配置示例已创建: secrets-example.md${NC}"
}

# 创建测试脚本
create_test_script() {
    echo -e "${BLUE}🧪 创建测试脚本...${NC}"
    
    cat > test-deployment.sh << 'EOF'
#!/bin/bash

# 测试部署脚本

echo "🧪 测试 Drone 部署流程"

# 检查 Drone 服务状态
echo "📊 检查 Drone 服务状态..."
docker-compose ps

# 测试推送代码
echo "📤 模拟代码推送..."
echo "// Test deployment $(date)" >> demo-web-app/main.go

git add .
git commit -m "Test: Trigger drone deployment $(date)"
git push origin main

echo "✅ 代码已推送，请在 Drone 界面查看构建状态"
EOF

    chmod +x test-deployment.sh
    echo -e "${GREEN}✅ 测试脚本已创建: test-deployment.sh${NC}"
}

# 显示后续步骤
show_next_steps() {
    echo -e "${GREEN}🎉 Drone 设置完成！${NC}"
    echo -e "${BLUE}📋 后续步骤：${NC}"
    echo "1. 访问 https://${DRONE_SERVER_HOST} 并使用 GitHub 登录"
    echo "2. 激活 ashwinyue/drone-demo 仓库"
    echo "3. 根据 secrets-example.md 添加必要的 Secrets"
    echo "4. 运行 ./test-deployment.sh 测试部署"
    echo "5. 查看详细文档: DRONE_SETUP_GUIDE.md"
    echo ""
    echo -e "${YELLOW}⚠️  注意事项：${NC}"
    echo "- 确保域名 ${DRONE_SERVER_HOST} 已正确解析到服务器"
    echo "- 如果使用 HTTPS，请配置 SSL 证书"
    echo "- 确保防火墙已开放 80 和 443 端口"
}

# 主函数
main() {
    echo -e "${BLUE}开始设置 Drone CI/CD...${NC}"
    
    check_docker
    check_kubectl
    get_user_input
    create_docker_compose
    start_drone
    create_secrets_example
    create_test_script
    show_next_steps
}

# 运行主函数
main