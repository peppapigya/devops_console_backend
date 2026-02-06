package helm

import (
	"devops-console-backend/pkg/configs"
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NewActionConfig 根据instance配置初始化Helm Action Configuration 通过instanceID动态切换K8s配置
func NewActionConfig(namespace string, instanceId uint) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	// 使用kubeconfig构建rest.Config
	config, exist := configs.GetK8sConfig(instanceId)
	if !exist {
		return nil, fmt.Errorf("k8s 未初始化")
	}

	// 初始化 ActionConfiguration
	err := actionConfig.Init(
		&restGetter{config: config, namespace: namespace},
		namespace,
		"secret", // Helm默认使用 secret 存储release信息
		func(format string, v ...interface{}) {
			log.Printf(format, v...)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	return actionConfig, nil
}

// restGetter 实现RESTClientGetter接口，用于Helm ActionConfig
type restGetter struct {
	config    *rest.Config
	namespace string
}

func (r *restGetter) ToRESTConfig() (*rest.Config, error) {
	return r.config, nil
}

func (r *restGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(r.config)
	if err != nil {
		return nil, err
	}
	// 2. 包装成内存缓存模式（Helm 需要 Cached 类型）
	return memory.NewMemCacheClient(dc), nil
}

func (r *restGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	// 获取所有 API 资源组
	gr, err := restmapper.GetAPIGroupResources(dc)
	// 构建并返回 Mapper
	return restmapper.NewDiscoveryRESTMapper(gr), nil
}

func (r *restGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(*api.NewConfig(), &clientcmd.ConfigOverrides{
		Context: api.Context{
			Namespace: r.namespace,
		},
	})
}
