# Go Drone Deploy - CI/CD 自动部署解决方案

这是一个基于 Drone CI/CD 的自动部署解决方案，包含了一个 Go 部署工具和示例 Web 应用。

## 🚀 快速开始

### 方式一：使用自动化脚本（推荐）

```bash
# 运行自动设置脚本
./setup-drone.sh
```

### 方式二：手动设置

1. **查看详细设置指南**
   ```bash
   cat DRONE_SETUP_GUIDE.md
   ```

2. **配置 GitHub OAuth 应用**
   - 访问 GitHub Settings > Developer settings > OAuth Apps
   - 创建新的 OAuth 应用
   - 获取 Client ID 和 Client Secret

3. **启动 Drone 服务**
   ```bash
   docker-compose up -d
   ```

## 📁 项目结构

```
├── demo-web-app/          # 示例 Web 应用
│   ├── .drone.yml         # Drone CI/CD 配置
│   ├── Dockerfile         # Docker 构建文件
│   ├── deploy-config.yaml # 部署配置
│   └── main.go           # 应用主文件
├── go-drone-deploy/       # 部署工具源码
│   ├── cmd/              # 命令行工具
│   ├── internal/         # 内部包
│   └── api/              # API 定义
├── setup-drone.sh         # 自动设置脚本
├── DRONE_SETUP_GUIDE.md   # 详细设置指南
└── README.md             # 本文件
```

## 🔧 核心功能

### Demo Web App
- ✅ 简单的 Go Web 服务
- ✅ 健康检查端点
- ✅ Docker 容器化
- ✅ Kubernetes 部署配置

### Go Drone Deploy 工具
- ✅ Docker 镜像构建和推送
- ✅ Kubernetes 自动部署
- ✅ 多环境支持（dev/staging/prod）
- ✅ Slack 通知集成
- ✅ 配置文件驱动

### Drone CI/CD 流水线
- ✅ 自动代码测试
- ✅ 自动构建应用
- ✅ 自动 Docker 镜像构建
- ✅ 自动 Kubernetes 部署
- ✅ 部署状态通知

## 🎯 部署流程

### 自动触发场景

| 操作 | 分支 | 触发结果 |
|------|------|----------|
| Push 代码 | `main` | 构建 + Docker 推送 |
| Push 代码 | `develop` | 完整部署到开发环境 |
| 创建 Tag | 任意 | 发布流水线 |
| Pull Request | 任意 | 仅测试和构建 |

### 手动部署命令

```bash
# 部署到开发环境
drone deploy ashwinyue/drone-demo <build_number> dev

# 部署到生产环境
drone deploy ashwinyue/drone-demo <build_number> prod
```

## 🔐 必需的 Secrets

在 Drone 界面中配置以下 Secrets：

```bash
# Docker Hub 认证
docker_username: your_dockerhub_username
docker_password: your_dockerhub_password

# Kubernetes 配置
kubeconfig: $(cat ~/.kube/config | base64 -w 0)
kubeconfig_dev: $(cat ~/.kube/config-dev | base64 -w 0)
kubeconfig_prod: $(cat ~/.kube/config-prod | base64 -w 0)

# 通知配置
webhook_url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
github_token: your_github_personal_access_token
```

## 🧪 测试部署

1. **推送代码测试**
   ```bash
   # 修改代码
   echo "// Test change $(date)" >> demo-web-app/main.go
   
   # 提交推送
   git add .
   git commit -m "Test deployment"
   git push origin main
   ```

2. **查看构建状态**
   - 访问 Drone 界面
   - 查看构建日志和状态

3. **验证部署结果**
   ```bash
   # 检查 Kubernetes 部署
   kubectl get pods -n default
   kubectl get services -n default
   
   # 查看应用日志
   kubectl logs -f deployment/demo-web-app
   ```

## 📊 监控和日志

### Drone 界面
- 构建历史和状态
- 实时构建日志
- 部署进度跟踪

### Kubernetes 监控
```bash
# 查看 Pod 状态
kubectl get pods -w

# 查看服务状态
kubectl get svc

# 查看应用日志
kubectl logs -f deployment/demo-web-app
```

### Slack 通知
- 部署开始通知
- 部署成功/失败通知
- 包含构建信息和链接

## 🔧 故障排除

### 常见问题

1. **构建失败**
   ```bash
   # 检查 Docker 权限
   sudo chmod 666 /var/run/docker.sock
   ```

2. **Kubernetes 部署失败**
   ```bash
   # 验证 kubeconfig
   kubectl cluster-info
   ```

3. **Secret 不可用**
   - 在 Drone 界面重新添加 Secret
   - 确保 Secret 名称正确

### 调试命令

```bash
# 查看 Drone 服务日志
docker logs drone-server
docker logs drone-runner

# 手动测试构建
docker run --rm -v $(pwd):/workspace -w /workspace golang:1.21-alpine go build -o demo-web-app ./demo-web-app
```

## 📚 相关文档

- [详细设置指南](DRONE_SETUP_GUIDE.md) - 完整的 Drone 设置步骤
- [Drone 官方文档](https://docs.drone.io/) - Drone CI/CD 官方文档
- [Kubernetes 部署指南](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) - K8s 部署文档

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。

---

**🎉 现在您可以通过推送代码到 GitHub 来触发自动部署了！**