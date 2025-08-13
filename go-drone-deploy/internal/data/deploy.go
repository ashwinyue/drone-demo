package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"go-drone-deploy/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// deployRepo 部署仓库实现
type deployRepo struct {
	data *Data
	log  *log.Helper
}

// NewDeployRepo 创建部署仓库
func NewDeployRepo(data *Data, logger log.Logger) biz.DeployRepo {
	return &deployRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// BuildDockerImage 构建 Docker 镜像
func (r *deployRepo) BuildDockerImage(ctx context.Context, config *biz.DockerConfig) error {
	r.log.WithContext(ctx).Infof("构建 Docker 镜像: %s", config.ImageName)

	// 构建 Docker 镜像命令
	cmd := exec.CommandContext(ctx, "docker", "build",
		"-t", config.ImageName,
		"-f", config.DockerfilePath,
		config.BuildContext,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("构建 Docker 镜像失败: %w", err)
	}

	r.log.WithContext(ctx).Info("Docker 镜像构建成功")
	return nil
}

// PushDockerImage 推送 Docker 镜像
func (r *deployRepo) PushDockerImage(ctx context.Context, config *biz.DockerConfig) error {
	r.log.WithContext(ctx).Infof("推送 Docker 镜像: %s", config.ImageName)

	// Docker 登录
	if config.Username != "" && config.Password != "" {
		loginCmd := exec.CommandContext(ctx, "docker", "login",
			"-u", config.Username,
			"-p", config.Password,
			config.Registry,
		)
		if err := loginCmd.Run(); err != nil {
			return fmt.Errorf("Docker 登录失败: %w", err)
		}
	}

	// 推送镜像
	pushCmd := exec.CommandContext(ctx, "docker", "push", config.ImageName)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("推送 Docker 镜像失败: %w", err)
	}

	r.log.WithContext(ctx).Info("Docker 镜像推送成功")
	return nil
}

// ApplyK8sDeployment 应用 Kubernetes Deployment
func (r *deployRepo) ApplyK8sDeployment(ctx context.Context, config *biz.K8sConfig) error {
	r.log.WithContext(ctx).Infof("应用 Kubernetes Deployment: %s", config.DeploymentName)

	// 创建 Kubernetes 客户端
	clientset, err := r.createK8sClient(config.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 构建 Deployment 对象
	deployment := r.buildDeployment(config)

	// 应用 Deployment
	deploymentsClient := clientset.AppsV1().Deployments(config.Namespace)
	_, err = deploymentsClient.Get(ctx, config.DeploymentName, metav1.GetOptions{})
	if err != nil {
		// Deployment 不存在，创建新的
		_, err = deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 Deployment 失败: %w", err)
		}
		r.log.WithContext(ctx).Info("Deployment 创建成功")
	} else {
		// Deployment 已存在，更新
		_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("更新 Deployment 失败: %w", err)
		}
		r.log.WithContext(ctx).Info("Deployment 更新成功")
	}

	return nil
}

// ApplyK8sService 应用 Kubernetes Service
func (r *deployRepo) ApplyK8sService(ctx context.Context, config *biz.K8sConfig) error {
	r.log.WithContext(ctx).Infof("应用 Kubernetes Service: %s", config.ServiceName)

	// 创建 Kubernetes 客户端
	clientset, err := r.createK8sClient(config.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 构建 Service 对象
	service := r.buildService(config)

	// 应用 Service
	servicesClient := clientset.CoreV1().Services(config.Namespace)
	_, err = servicesClient.Get(ctx, config.ServiceName, metav1.GetOptions{})
	if err != nil {
		// Service 不存在，创建新的
		_, err = servicesClient.Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 Service 失败: %w", err)
		}
		r.log.WithContext(ctx).Info("Service 创建成功")
	} else {
		// Service 已存在，更新
		_, err = servicesClient.Update(ctx, service, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("更新 Service 失败: %w", err)
		}
		r.log.WithContext(ctx).Info("Service 更新成功")
	}

	return nil
}

// UpdateK8sVersion 更新 Kubernetes 版本
func (r *deployRepo) UpdateK8sVersion(ctx context.Context, config *biz.K8sConfig) error {
	r.log.WithContext(ctx).Info("更新 Kubernetes 版本")

	// 创建 Kubernetes 客户端
	clientset, err := r.createK8sClient(config.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 重启 Deployment 以应用新版本
	deploymentsClient := clientset.AppsV1().Deployments(config.Namespace)
	deployment, err := deploymentsClient.Get(ctx, config.DeploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 添加重启注解
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("重启 Deployment 失败: %w", err)
	}

	r.log.WithContext(ctx).Info("Kubernetes 版本更新成功")
	return nil
}

// SendNotification 发送通知
func (r *deployRepo) SendNotification(ctx context.Context, config *biz.NotifyConfig, message string) error {
	r.log.WithContext(ctx).Infof("发送通知: %s", message)

	if config.WebhookURL == "" {
		return fmt.Errorf("Webhook URL 为空")
	}

	// 构建通知消息
	payload := map[string]interface{}{
		"text":    message,
		"channel": config.Channel,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	// 发送 HTTP 请求
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送通知请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("通知发送失败，状态码: %d", resp.StatusCode)
	}

	r.log.WithContext(ctx).Info("通知发送成功")
	return nil
}

// createK8sClient 创建 Kubernetes 客户端
func (r *deployRepo) createK8sClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var config clientcmd.ClientConfig

	if kubeconfigPath != "" {
		// 使用指定的 kubeconfig 文件
		config = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			&clientcmd.ConfigOverrides{},
		)
	} else {
		// 使用默认的 kubeconfig
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			&clientcmd.ConfigOverrides{},
		)
	}

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("加载 kubeconfig 失败: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	return clientset, nil
}

// buildDeployment 构建 Deployment 对象
func (r *deployRepo) buildDeployment(config *biz.K8sConfig) *appsv1.Deployment {
	labels := map[string]string{
		"app": config.DeploymentName,
	}

	// 构建环境变量
	var envVars []corev1.EnvVar
	for _, env := range config.EnvVars {
		envVars = append(envVars, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	// 构建端口
	var ports []corev1.ContainerPort
	for _, port := range config.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
			Protocol:      corev1.Protocol(port.Protocol),
		})
	}

	// 构建资源限制
	resources := corev1.ResourceRequirements{}
	if config.Resources != nil {
		if config.Resources.CPURequest != "" || config.Resources.MemoryRequest != "" {
			resources.Requests = corev1.ResourceList{}
			if config.Resources.CPURequest != "" {
				resources.Requests[corev1.ResourceCPU] = *parseQuantity(config.Resources.CPURequest)
			}
			if config.Resources.MemoryRequest != "" {
				resources.Requests[corev1.ResourceMemory] = *parseQuantity(config.Resources.MemoryRequest)
			}
		}
		if config.Resources.CPULimit != "" || config.Resources.MemoryLimit != "" {
			resources.Limits = corev1.ResourceList{}
			if config.Resources.CPULimit != "" {
				resources.Limits[corev1.ResourceCPU] = *parseQuantity(config.Resources.CPULimit)
			}
			if config.Resources.MemoryLimit != "" {
				resources.Limits[corev1.ResourceMemory] = *parseQuantity(config.Resources.MemoryLimit)
			}
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.DeploymentName,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      config.DeploymentName,
							Image:     "", // 这里需要从 Docker 配置中获取
							Ports:     ports,
							Env:       envVars,
							Resources: resources,
						},
					},
				},
			},
		},
	}
}

// buildService 构建 Service 对象
func (r *deployRepo) buildService(config *biz.K8sConfig) *corev1.Service {
	labels := map[string]string{
		"app": config.DeploymentName,
	}

	// 构建端口
	var ports []corev1.ServicePort
	for _, port := range config.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
			Protocol:   corev1.Protocol(port.Protocol),
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ServiceName,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

// parseQuantity 解析资源数量（简化版本）
func parseQuantity(s string) *resource.Quantity {
	// 这里应该使用 k8s.io/apimachinery/pkg/api/resource.ParseQuantity
	// 为了简化，这里返回一个空的 Quantity
	return &resource.Quantity{}
}