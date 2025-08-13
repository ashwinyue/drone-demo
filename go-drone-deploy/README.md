# Go Drone Deploy

基于 Kratos 框架构建的现代化 Drone CI/CD 部署工具，专为 Go 项目设计。

## 特性

- 🚀 **现代化架构**: 基于 Kratos 微服务框架
- 🐳 **Docker 支持**: 自动构建和推送 Docker 镜像
- ☸️ **Kubernetes 集成**: 原生支持 K8s 部署
- 🔔 **通知系统**: 支持 Webhook 通知
- 🛠️ **灵活配置**: 支持多环境配置
- 📊 **完整日志**: 详细的部署日志记录

## 快速开始

### 安装

```bash
# 克隆项目
git clone https://github.com/solariswu/go-drone-deploy.git
cd go-drone-deploy

# 安装依赖
go mod tidy

# 生成代码
go generate ./...

# 构建
go build -o ./bin/ ./...
```

### 配置

编辑 `configs/config.yaml` 文件，配置你的部署参数：

```yaml
deploy:
  project_name: "your-app"
  author: "your-team"
  namespace: "default"
  version: "v1.0.0"
  env: "dev"
  
  docker:
    registry: "docker.io"
    username: "your-username"
    password: "your-password"
    image_name: "your-app:latest"
    dockerfile_path: "./Dockerfile"
    build_context: "."
  
  k8s:
    kubeconfig_path: "~/.kube/config"
    namespace: "default"
    deployment_name: "your-app"
    service_name: "your-app-service"
    replicas: 3
```

### 使用

```bash
# 完整部署流程
./bin/go-drone-deploy -conf ./configs/config.yaml -flow all -env prod

# 仅构建和推送 Docker 镜像
./bin/go-drone-deploy -flow docker

# 仅部署到 Kubernetes
./bin/go-drone-deploy -flow k8s

# 标准流程（不包含通知）
./bin/go-drone-deploy -flow standard

# 显示版本信息
./bin/go-drone-deploy -version
```

## 部署流程

### 支持的流程类型

- `all`: 完整部署流程（Docker + K8s + 通知）
- `docker`: 仅 Docker 构建和推送
- `k8s`: 仅 Kubernetes 部署
- `standard`: 标准流程（Docker + K8s，不包含通知）
- `notify`: 仅发送通知

### 流程说明

1. **Docker 阶段**
   - 构建 Docker 镜像
   - 推送到镜像仓库

2. **Kubernetes 阶段**
   - 创建/更新 Deployment
   - 创建/更新 Service
   - 滚动更新应用版本

3. **通知阶段**
   - 发送部署成功通知
   - 支持 Webhook 集成

## Drone CI 集成

在你的项目根目录创建 `.drone.yml` 文件：

```yaml
kind: pipeline
type: docker
name: deploy

steps:
- name: deploy
  image: solariswu/go-drone-deploy:latest
  settings:
    config: ./configs/config.yaml
    flow: all
    env: production
  when:
    branch:
    - main
```

## 本地开发

### 使用 Kind 进行本地测试

```bash
# 创建 Kind 集群
kind create cluster --name drone-deploy-test

# 设置 kubectl 上下文
kubectl cluster-info --context kind-drone-deploy-test

# 测试部署
./bin/go-drone-deploy -flow k8s -env dev
```

### 开发环境设置

```bash
# 安装开发依赖
go install github.com/google/wire/cmd/wire@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest

# 生成代码
go generate ./...

# 运行测试
go test ./...
```

## 配置参考

### Docker 配置

| 参数 | 说明 | 示例 |
|------|------|------|
| `registry` | Docker 镜像仓库 | `docker.io` |
| `username` | 仓库用户名 | `your-username` |
| `password` | 仓库密码 | `your-password` |
| `image_name` | 镜像名称 | `app:latest` |
| `dockerfile_path` | Dockerfile 路径 | `./Dockerfile` |
| `build_context` | 构建上下文 | `.` |

### Kubernetes 配置

| 参数 | 说明 | 示例 |
|------|------|------|
| `kubeconfig_path` | kubeconfig 文件路径 | `~/.kube/config` |
| `namespace` | K8s 命名空间 | `default` |
| `deployment_name` | Deployment 名称 | `app` |
| `service_name` | Service 名称 | `app-service` |
| `replicas` | 副本数量 | `3` |

### 通知配置

| 参数 | 说明 | 示例 |
|------|------|------|
| `enabled` | 是否启用通知 | `true` |
| `webhook_url` | Webhook URL | `https://hooks.slack.com/...` |
| `channel` | 通知频道 | `#deployment` |

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

