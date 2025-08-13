package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-drone-deploy/internal/biz"
	"go-drone-deploy/internal/data"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "go-drone-deploy"
	// Version is the version of the compiled software.
	Version = "v1.0.0"
	// flagconf is the config flag.
	flagconf string
	// flagflow is the deploy flow flag.
	flagflow string
	// flagenv is the environment flag.
	flagenv string
	// flagversion shows version info.
	flagversion bool
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs/config.yaml", "配置文件路径")
	flag.StringVar(&flagflow, "flow", "all", "部署流程 (all/docker/k8s/notify/standard)")
	flag.StringVar(&flagenv, "env", "dev", "部署环境 (dev/staging/prod)")
	flag.BoolVar(&flagversion, "version", false, "显示版本信息")
}

func main() {
	flag.Parse()

	// 显示版本信息
	if flagversion {
		fmt.Printf("%s %s\n", Name, Version)
		return
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	go handleSignals(cancel)

	// 创建日志器
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", Name,
		"service.version", Version,
	)

	// 加载配置
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		log.NewHelper(logger).Fatalf("加载配置失败: %v", err)
	}

	// 创建简单的部署配置（示例）
	deployConfig := &biz.DeployConfig{
		ProjectName: "example-app",
		Author:      "drone-deploy",
		Namespace:   "default",
		Version:     "latest",
		Env:         flagenv,
		Docker: &biz.DockerConfig{
			Registry:       "docker.io",
			ImageName:      "example-app:latest",
			DockerfilePath: "./Dockerfile",
			BuildContext:   ".",
		},
		K8s: &biz.K8sConfig{
			Namespace:      "default",
			DeploymentName: "example-app",
			ServiceName:    "example-app-service",
			Replicas:       1,
			Ports: []*biz.Port{
				{
					Name:       "http",
					Port:       80,
					TargetPort: 8080,
					Protocol:   "TCP",
				},
			},
		},
	}

	// 创建数据层和业务层
	dataRepo, cleanup, err := data.NewData(nil, logger)
	if err != nil {
		log.NewHelper(logger).Fatalf("创建数据层失败: %v", err)
	}
	defer cleanup()
	
	deployRepo := data.NewDeployRepo(dataRepo, logger)
	deployUC := biz.NewDeployUsecase(deployRepo, logger)

	// 映射流程名称
	var flow biz.DeployFlow
	switch flagflow {
	case "full":
		flow = biz.FlowAll
	case "docker":
		flow = biz.FlowDocker
	case "k8s":
		flow = biz.FlowK8s
	case "notify":
		flow = biz.FlowNotify
	case "standard":
		flow = biz.FlowStandard
	default:
		log.NewHelper(logger).Fatalf("不支持的部署流程: %s", flagflow)
	}

	// 执行部署
	if err := deployUC.Deploy(ctx, deployConfig, flow); err != nil {
		log.NewHelper(logger).Fatalf("部署失败: %v", err)
	}

	log.NewHelper(logger).Info("部署完成")
}

func handleSignals(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cancel()
}
