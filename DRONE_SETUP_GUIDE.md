# Drone CI/CD 部署设置指南

本指南将帮助您设置 Drone CI/CD 流水线，实现 demo-web-app 项目的自动部署。

## 前置条件

1. **Drone 服务器已安装并运行**
2. **Docker 环境可用**
3. **Kubernetes 集群可访问**
4. **GitHub 仓库已连接到 Drone**

## 1. Drone 服务器设置

### 1.1 安装 Drone 服务器

```bash
# 使用 Docker Compose 安装 Drone
docker run \
  --volume=/var/lib/drone:/data \
  --env=DRONE_GITHUB_CLIENT_ID=your_github_client_id \
  --env=DRONE_GITHUB_CLIENT_SECRET=your_github_client_secret \
  --env=DRONE_RPC_SECRET=your_rpc_secret \
  --env=DRONE_SERVER_HOST=your_drone_server_host \
  --env=DRONE_SERVER_PROTO=https \
  --publish=80:80 \
  --publish=443:443 \
  --restart=always \
  --detach=true \
  --name=drone \
  drone/drone:2
```

### 1.2 安装 Drone Runner

```bash
docker run --detach \
  --volume=/var/run/docker.sock:/var/run/docker.sock \
  --env=DRONE_RPC_PROTO=https \
  --env=DRONE_RPC_HOST=your_drone_server_host \
  --env=DRONE_RPC_SECRET=your_rpc_secret \
  --env=DRONE_RUNNER_CAPACITY=2 \
  --env=DRONE_RUNNER_NAME=my-first-runner \
  --publish=3000:3000 \
  --restart=always \
  --name=runner \
  drone/drone-runner-docker:1
```

## 2. GitHub 集成设置

### 2.1 创建 GitHub OAuth 应用

1. 访问 GitHub Settings > Developer settings > OAuth Apps
2. 点击 "New OAuth App"
3. 填写应用信息：
   - Application name: `Drone CI`
   - Homepage URL: `https://your-drone-server.com`
   - Authorization callback URL: `https://your-drone-server.com/login`
4. 获取 Client ID 和 Client Secret

### 2.2 激活仓库

1. 访问 Drone 界面：`https://your-drone-server.com`
2. 使用 GitHub 账号登录
3. 在仓库列表中找到 `ashwinyue/drone-demo`
4. 点击 "ACTIVATE" 激活仓库

## 3. 配置 Secrets

在 Drone 界面中为仓库添加以下 Secrets：

### 3.1 Docker 相关 Secrets

```bash
# Docker Hub 用户名
docker_username: your_dockerhub_username

# Docker Hub 密码或访问令牌
docker_password: your_dockerhub_password
```

### 3.2 Kubernetes 相关 Secrets

```bash
# 开发环境 kubeconfig（base64 编码）
kubeconfig_dev: $(cat ~/.kube/config | base64 -w 0)

# 生产环境 kubeconfig（base64 编码）
kubeconfig_prod: $(cat ~/.kube/config-prod | base64 -w 0)

# 通用 kubeconfig
kubeconfig: $(cat ~/.kube/config | base64 -w 0)
```

### 3.3 通知相关 Secrets

```bash
# Slack Webhook URL
webhook_url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# GitHub Token（用于发布）
github_token: your_github_personal_access_token
```

## 4. 部署流程说明

### 4.1 自动触发场景

| 事件 | 分支 | 触发的步骤 | 说明 |
|------|------|------------|------|
| Push | main | test, build, docker-deploy | 主分支推送，构建并推送 Docker 镜像 |
| Push | develop | test, build, full-deploy-dev | 开发分支推送，完整部署到开发环境 |
| Pull Request | 任意 | test, build | 仅运行测试和构建 |
| Tag | 任意 | release pipeline | 发布流水线，构建发布版本 |

### 4.2 手动触发场景

```bash
# 部署到开发环境
drone deploy ashwinyue/drone-demo <build_number> dev

# 部署到预发布环境
drone deploy ashwinyue/drone-demo <build_number> staging

# 部署到生产环境
drone deploy ashwinyue/drone-demo <build_number> prod
```

## 5. 验证部署

### 5.1 检查 Drone 界面

1. 访问 `https://your-drone-server.com`
2. 进入 `ashwinyue/drone-demo` 仓库
3. 查看构建历史和状态

### 5.2 推送代码测试

```bash
# 克隆仓库
git clone https://github.com/ashwinyue/drone-demo.git
cd drone-demo

# 修改代码
echo "// Test change" >> demo-web-app/main.go

# 提交并推送
git add .
git commit -m "Test drone deployment"
git push origin main
```

### 5.3 查看部署结果

```bash
# 检查 Kubernetes 部署
kubectl get pods -n default
kubectl get services -n default

# 检查应用日志
kubectl logs -f deployment/demo-web-app -n default
```

## 6. 故障排除

### 6.1 常见问题

**问题：构建失败，提示权限错误**
```bash
# 解决方案：检查 Docker socket 权限
sudo chmod 666 /var/run/docker.sock
```

**问题：Kubernetes 部署失败**
```bash
# 解决方案：验证 kubeconfig
echo $KUBECONFIG | base64 -d > /tmp/kubeconfig
kubectl --kubeconfig=/tmp/kubeconfig get nodes
```

**问题：Secret 不可用**
```bash
# 解决方案：重新添加 Secret
# 在 Drone 界面中删除并重新添加相关 Secret
```

### 6.2 调试命令

```bash
# 查看 Drone 服务器日志
docker logs drone

# 查看 Runner 日志
docker logs runner

# 手动运行构建步骤
docker run --rm -v $(pwd):/workspace -w /workspace golang:1.21-alpine go build -o demo-web-app .
```

## 7. 高级配置

### 7.1 多环境部署

可以通过修改 `.drone.yml` 文件来支持更多环境：

```yaml
# 添加测试环境部署
- name: deploy-test
  image: solariswu/go-drone-deploy:latest
  environment:
    KUBECONFIG:
      from_secret: kubeconfig_test
  commands:
    - /go-drone-deploy -flow=k8s -env=test
  when:
    event: [promote]
    target: [test]
```

### 7.2 条件部署

```yaml
# 仅在特定文件变更时部署
- name: deploy-frontend
  image: solariswu/go-drone-deploy:latest
  commands:
    - /go-drone-deploy -flow=k8s -env=dev
  when:
    changeset:
      includes:
        - "frontend/**"
        - "*.js"
```

## 8. 监控和通知

### 8.1 Slack 通知设置

1. 创建 Slack Webhook
2. 在 Drone Secrets 中添加 `webhook_url`
3. 部署完成后会自动发送通知

### 8.2 邮件通知

可以添加邮件通知步骤：

```yaml
- name: email-notify
  image: drillster/drone-email
  settings:
    host: smtp.gmail.com
    username:
      from_secret: email_username
    password:
      from_secret: email_password
    from: drone@yourcompany.com
    recipients:
      - team@yourcompany.com
  when:
    status: [success, failure]
```

## 总结

完成以上设置后，您的 Drone CI/CD 流水线将能够：

1. ✅ 自动检测代码推送
2. ✅ 运行测试和构建
3. ✅ 构建并推送 Docker 镜像
4. ✅ 部署到 Kubernetes 集群
5. ✅ 发送部署通知
6. ✅ 支持多环境部署

现在您可以通过推送代码到 GitHub 来触发自动部署了！