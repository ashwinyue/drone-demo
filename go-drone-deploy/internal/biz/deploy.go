package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// DeployFlow 定义部署流程类型
type DeployFlow string

const (
	FlowAll      DeployFlow = "all"      // 完整部署流程
	FlowDocker   DeployFlow = "docker"   // 仅 Docker 构建和推送
	FlowK8s      DeployFlow = "k8s"      // 仅 Kubernetes 部署
	FlowNotify   DeployFlow = "notify"   // 仅通知
	FlowStandard DeployFlow = "standard" // 标准流程（不包含通知）
)

// DeployConfig 部署配置
type DeployConfig struct {
	ProjectName string
	Author      string
	Namespace   string
	Version     string
	Env         string
	Docker      *DockerConfig
	K8s         *K8sConfig
	Notify      *NotifyConfig
}

// DockerConfig Docker 配置
type DockerConfig struct {
	Registry       string
	Username       string
	Password       string
	ImageName      string
	DockerfilePath string
	BuildContext   string
}

// K8sConfig Kubernetes 配置
type K8sConfig struct {
	KubeconfigPath string
	Namespace      string
	DeploymentName string
	ServiceName    string
	Replicas       int32
	Resources      *Resources
	Ports          []*Port
	EnvVars        []*EnvVar
}

// Resources 资源配置
type Resources struct {
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
}

// Port 端口配置
type Port struct {
	Name       string
	Port       int32
	TargetPort int32
	Protocol   string
}

// EnvVar 环境变量
type EnvVar struct {
	Name  string
	Value string
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	Enabled    bool
	WebhookURL string
	Channel    string
}

// DeployRepo 部署仓库接口
type DeployRepo interface {
	// Docker 相关
	BuildDockerImage(ctx context.Context, config *DockerConfig) error
	PushDockerImage(ctx context.Context, config *DockerConfig) error
	
	// Kubernetes 相关
	ApplyK8sDeployment(ctx context.Context, config *K8sConfig) error
	ApplyK8sService(ctx context.Context, config *K8sConfig) error
	UpdateK8sVersion(ctx context.Context, config *K8sConfig) error
	
	// 通知相关
	SendNotification(ctx context.Context, config *NotifyConfig, message string) error
}

// DeployUsecase 部署用例
type DeployUsecase struct {
	repo DeployRepo
	log  *log.Helper
}

// NewDeployUsecase 创建部署用例
func NewDeployUsecase(repo DeployRepo, logger log.Logger) *DeployUsecase {
	return &DeployUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Deploy 执行部署
func (uc *DeployUsecase) Deploy(ctx context.Context, config *DeployConfig, flow DeployFlow) error {
	uc.log.WithContext(ctx).Infof("开始部署，项目: %s, 环境: %s, 流程: %s", config.ProjectName, config.Env, flow)

	switch flow {
	case FlowAll:
		return uc.deployAll(ctx, config)
	case FlowDocker:
		return uc.deployDocker(ctx, config)
	case FlowK8s:
		return uc.deployK8s(ctx, config)
	case FlowNotify:
		return uc.deployNotify(ctx, config)
	case FlowStandard:
		return uc.deployStandard(ctx, config)
	default:
		return fmt.Errorf("不支持的部署流程: %s", flow)
	}
}

// deployAll 完整部署流程
func (uc *DeployUsecase) deployAll(ctx context.Context, config *DeployConfig) error {
	// 1. Docker 构建和推送
	if err := uc.deployDocker(ctx, config); err != nil {
		return fmt.Errorf("Docker 部署失败: %w", err)
	}

	// 2. Kubernetes 部署
	if err := uc.deployK8s(ctx, config); err != nil {
		return fmt.Errorf("Kubernetes 部署失败: %w", err)
	}

	// 3. 发送通知
	if err := uc.deployNotify(ctx, config); err != nil {
		uc.log.WithContext(ctx).Warnf("通知发送失败: %v", err)
	}

	return nil
}

// deployStandard 标准部署流程（不包含通知）
func (uc *DeployUsecase) deployStandard(ctx context.Context, config *DeployConfig) error {
	// 1. Docker 构建和推送
	if err := uc.deployDocker(ctx, config); err != nil {
		return fmt.Errorf("Docker 部署失败: %w", err)
	}

	// 2. Kubernetes 部署
	if err := uc.deployK8s(ctx, config); err != nil {
		return fmt.Errorf("Kubernetes 部署失败: %w", err)
	}

	return nil
}

// deployDocker Docker 构建和推送
func (uc *DeployUsecase) deployDocker(ctx context.Context, config *DeployConfig) error {
	if config.Docker == nil {
		return fmt.Errorf("Docker 配置为空")
	}

	uc.log.WithContext(ctx).Info("开始构建 Docker 镜像")
	if err := uc.repo.BuildDockerImage(ctx, config.Docker); err != nil {
		return fmt.Errorf("构建 Docker 镜像失败: %w", err)
	}

	uc.log.WithContext(ctx).Info("开始推送 Docker 镜像")
	if err := uc.repo.PushDockerImage(ctx, config.Docker); err != nil {
		return fmt.Errorf("推送 Docker 镜像失败: %w", err)
	}

	return nil
}

// deployK8s Kubernetes 部署
func (uc *DeployUsecase) deployK8s(ctx context.Context, config *DeployConfig) error {
	if config.K8s == nil {
		return fmt.Errorf("Kubernetes 配置为空")
	}

	uc.log.WithContext(ctx).Info("开始部署 Kubernetes Deployment")
	if err := uc.repo.ApplyK8sDeployment(ctx, config.K8s); err != nil {
		return fmt.Errorf("部署 Kubernetes Deployment 失败: %w", err)
	}

	uc.log.WithContext(ctx).Info("开始部署 Kubernetes Service")
	if err := uc.repo.ApplyK8sService(ctx, config.K8s); err != nil {
		return fmt.Errorf("部署 Kubernetes Service 失败: %w", err)
	}

	uc.log.WithContext(ctx).Info("开始更新 Kubernetes 版本")
	if err := uc.repo.UpdateK8sVersion(ctx, config.K8s); err != nil {
		return fmt.Errorf("更新 Kubernetes 版本失败: %w", err)
	}

	return nil
}

// deployNotify 发送通知
func (uc *DeployUsecase) deployNotify(ctx context.Context, config *DeployConfig) error {
	if config.Notify == nil || !config.Notify.Enabled {
		uc.log.WithContext(ctx).Info("通知功能未启用")
		return nil
	}

	message := fmt.Sprintf("项目 %s 在 %s 环境部署成功，版本: %s", config.ProjectName, config.Env, config.Version)
	return uc.repo.SendNotification(ctx, config.Notify, message)
}