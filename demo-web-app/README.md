# Demo Web App
test push
这是一个用于演示 [go-drone-deploy](../go-drone-deploy) 工具的简单 Go web 应用程序。

## 🚀 功能特性

- 简单的 HTTP API 服务
- 健康检查端点
- 应用信息端点
- 容器化支持
- Kubernetes 部署配置
- Drone CI/CD 自动部署

## 📋 API 端点

- `GET /` - 主页面，返回应用基本信息
- `GET /health` - 健康检查，返回应用状态和运行时间
- `GET /api/info` - 应用详细信息，包括版本、环境等

## 🏃‍♂️ 本地运行

### 直接运行

```bash
# 克隆项目
git clone <repository-url>
cd demo-web-app

# 安装依赖
go mod download

# 运行应用
go run main.go

# 或者构建后运行
go build -o demo-web-app .
./demo-web-app
```

应用将在 `http://localhost:8080` 启动。

### 使用 Docker 运行

```bash
# 构建 Docker 镜像
docker build -t demo-web-app:latest .

# 运行容器
docker run -p 8080:8080 demo-web-app:latest
```

## 🐳 Docker 部署

项目包含了完整的 Dockerfile，支持多阶段构建：

```dockerfile
# 查看 Dockerfile
cat Dockerfile
```

## ☸️ Kubernetes 部署

项目在 `k8s/` 目录下包含了 Kubernetes 部署配置：

- `deployment.yaml` - 应用部署配置
- `service.yaml` - 服务配置（包含 ClusterIP 和 NodePort）

```bash
# 部署到 Kubernetes
kubectl apply -f k8s/

# 查看部署状态
kubectl get pods -l app=demo-web-app
kubectl get services -l app=demo-web-app

# 访问应用（NodePort 方式）
# 应用将在节点的 30080 端口可访问
curl http://<node-ip>:30080
```

## 🚁 Drone CI/CD 部署

这个项目配置了完整的 Drone CI/CD 流水线，使用 `go-drone-deploy` 工具进行自动部署。

### 流水线配置

项目包含 `.drone.yml` 配置文件，定义了以下流水线：

1. **测试阶段** - 运行单元测试和代码检查
2. **构建阶段** - 编译 Go 应用
3. **Docker 部署** - 使用 go-drone-deploy 构建和推送 Docker 镜像
4. **Kubernetes 部署** - 使用 go-drone-deploy 部署到 K8s 集群
5. **完整部署** - 执行完整的部署流程（Docker + K8s + 通知）
6. **通知** - 发送部署结果通知

### 触发条件

- **推送到 main/develop 分支** - 触发 Docker 构建和推送
- **推送到 develop 分支** - 触发开发环境完整部署
- **Promote 事件** - 触发指定环境部署
- **创建 Tag** - 触发发布流水线

### 环境变量配置

在 Drone 中需要配置以下 Secret：

```bash
# Docker 相关
docker_username     # Docker Hub 用户名
docker_password     # Docker Hub 密码

# Kubernetes 相关
kubeconfig          # Kubernetes 配置文件
kubeconfig_dev      # 开发环境 K8s 配置
kubeconfig_prod     # 生产环境 K8s 配置

# 通知相关
webhook_url         # Slack/钉钉等 Webhook URL

# 发布相关
github_token        # GitHub Token（用于发布）
```

### 部署流程示例

1. **开发环境自动部署**：
   ```bash
   # 推送代码到 develop 分支
   git push origin develop
   
   # Drone 将自动执行：
   # 1. 测试和构建
   # 2. Docker 镜像构建和推送
   # 3. Kubernetes 部署
   # 4. 发送通知
   ```

2. **生产环境手动部署**：
   ```bash
   # 在 Drone UI 中手动触发 Promote 到 prod 环境
   # 或使用 Drone CLI
   drone build promote <repo> <build> prod
   ```

3. **版本发布**：
   ```bash
   # 创建并推送 tag
   git tag v1.0.1
   git push origin v1.0.1
   
   # 将触发发布流水线，创建 GitHub Release
   ```

## 🔧 go-drone-deploy 配置

项目包含 `deploy-config.yaml` 配置文件，定义了 go-drone-deploy 工具的部署参数：

- Docker 配置（镜像名称、注册表等）
- Kubernetes 配置（命名空间、副本数、资源限制等）
- 通知配置（Webhook URL、频道等）

可以根据实际环境修改配置文件中的参数。

## 📝 使用说明

1. **Fork 这个项目**到你的 Git 仓库
2. **配置 Drone CI**，连接你的 Git 仓库
3. **设置必要的 Secret**（Docker、Kubernetes、通知等）
4. **修改配置文件**（deploy-config.yaml、.drone.yml）适应你的环境
5. **推送代码**，观察 Drone 自动执行部署流程

## 🎯 测试部署

```bash
# 测试应用是否正常运行
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/info

# 预期响应示例
{
  "message": "Hello from Demo Web App! 🚀",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "v1.0.0",
  "env": "development",
  "hostname": "localhost"
}
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个演示项目！

## 📄 许可证

MIT License