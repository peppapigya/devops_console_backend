package configs

import (
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	k8sClients     map[uint]*kubernetes.Clientset
	k8sClientsLock sync.RWMutex
	configMap      map[uint]*rest.Config
	k8sConfigLock  sync.RWMutex
)

// InitK8sClients 初始化所有K8s类型的客户端
func InitK8sClients() error {
	k8sClientsLock.Lock()
	defer k8sClientsLock.Unlock()

	k8sClients = make(map[uint]*kubernetes.Clientset)
	configMap = make(map[uint]*rest.Config)

	// 查询所有kubernetes类型的实例
	instanceRepo := NewInstanceRepository()
	authConfigRepo := NewAuthConfigRepository()

	// 先获取kubernetes实例类型
	instanceTypeRepo := NewInstanceTypeRepository()
	k8sType, err := instanceTypeRepo.GetByName("kubernetes")
	if err != nil {
		logs.Error(map[string]interface{}{
			"error": err.Error(),
		}, "查询kubernetes实例类型失败")
		return err
	}

	instances, err := instanceRepo.GetByTypeID(k8sType.ID)
	if err != nil {
		logs.Error(map[string]interface{}{
			"error": err.Error(),
		}, "查询kubernetes实例失败")
		return err
	}

	for _, instance := range instances {
		if instance.Status != "active" {
			continue
		}

		// 获取认证配置
		authConfigs, err := authConfigRepo.GetByInstanceID(instance.ID)
		if err != nil {
			logs.Warning(map[string]interface{}{
				"instance_id": instance.ID,
				"error":       err.Error(),
			}, "获取实例认证配置失败")
			continue
		}

		if len(authConfigs) == 0 {
			logs.Warning(map[string]interface{}{
				"instance_id": instance.ID,
			}, "实例没有认证配置")
			continue
		}

		authConfig := authConfigs[0]
		if authConfig.AuthType != "kubeconfig" {
			logs.Warning(map[string]interface{}{
				"instance_id": instance.ID,
				"auth_type":   authConfig.AuthType,
			}, "不支持的k8s认证类型")
			continue
		}

		// 解析kubeconfig内容（可能是JSON格式）
		var kubeconfigContent string
		if strings.HasPrefix(authConfig.ConfigValue, "{") {
			// JSON格式，需要解析
			var configData map[string]interface{}
			if err := json.Unmarshal([]byte(authConfig.ConfigValue), &configData); err == nil {
				if kubeconfig, ok := configData["kubeconfigContent"].(string); ok {
					kubeconfigContent = kubeconfig
				}
			}
		} else {
			// 直接的kubeconfig内容
			kubeconfigContent = authConfig.ConfigValue
		}

		if kubeconfigContent == "" {
			logs.Warning(map[string]interface{}{
				"instance_id": instance.ID,
			}, "kubeconfig内容为空")
			continue
		}

		// 初始化客户端
		restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
		if err != nil {
			logs.Error(map[string]interface{}{
				"instance_id": instance.ID,
				"error":       err.Error(),
			}, "构建k8s配置失败")
			continue
		}
		configMap[instance.ID] = restConfig
		clientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			logs.Error(map[string]interface{}{
				"instance_id": instance.ID,
				"error":       err.Error(),
			}, "创建k8s客户端失败")
			continue
		}
		k8sClients[instance.ID] = clientSet
		logs.Info(map[string]interface{}{
			"instance_id":   instance.ID,
			"instance_name": instance.Name,
		}, "k8s客户端初始化成功")
	}

	logs.Info(map[string]interface{}{
		"count": len(k8sClients),
	}, "k8s客户端初始化完成")

	return nil
}

// GetK8sClient 获取指定实例的K8s客户端
func GetK8sClient(instanceID uint) (*kubernetes.Clientset, bool) {
	k8sClientsLock.RLock()
	defer k8sClientsLock.RUnlock()

	client, exists := k8sClients[instanceID]
	return client, exists
}

func GetK8sConfig(instanceID uint) (*rest.Config, bool) {
	k8sConfigLock.RLock()
	defer k8sConfigLock.RUnlock()
	config, exists := configMap[instanceID]
	return config, exists
}

// AddK8sClient 添加新的K8s客户端
func AddK8sClient(instance *dal.Instance, authConfig *dal.AuthConfig) error {
	k8sClientsLock.Lock()
	defer k8sClientsLock.Unlock()

	if authConfig.AuthType != "kubeconfig" {
		return fmt.Errorf("不支持的k8s认证类型: %s", authConfig.AuthType)
	}

	// 解析kubeconfig内容（可能是JSON格式）
	var kubeconfigContent string
	if strings.HasPrefix(authConfig.ConfigValue, "{") {
		// JSON格式，需要解析
		var configData map[string]interface{}
		if err := json.Unmarshal([]byte(authConfig.ConfigValue), &configData); err == nil {
			if kubeconfig, ok := configData["kubeconfigContent"].(string); ok {
				kubeconfigContent = kubeconfig
			}
		}
	} else {
		// 直接的kubeconfig内容
		kubeconfigContent = authConfig.ConfigValue
	}

	if kubeconfigContent == "" {
		return fmt.Errorf("kubeconfig内容为空")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	if k8sClients == nil {
		k8sClients = make(map[uint]*kubernetes.Clientset)
	}
	k8sClients[instance.ID] = clientSet
	logs.Info(map[string]interface{}{
		"instance_id":   instance.ID,
		"instance_name": instance.Name,
	}, "k8s客户端添加成功")

	return nil
}

// RemoveK8sClient 移除K8s客户端
func RemoveK8sClient(instanceID uint) {
	k8sClientsLock.Lock()
	defer k8sClientsLock.Unlock()

	delete(k8sClients, instanceID)
	logs.Info(map[string]interface{}{
		"instance_id": instanceID,
	}, "k8s客户端移除成功")
}

// GetMetricsClient 获取metrics客户端
func GetMetricsClient(instanceID uint) (*metricsv.Clientset, bool) {
	k8sClientsLock.RLock()
	_, exists := k8sClients[instanceID]
	k8sClientsLock.RUnlock()

	if !exists {
		return nil, false
	}

	// 获取认证配置
	authConfigRepo := NewAuthConfigRepository()
	authConfigs, err := authConfigRepo.GetByInstanceID(instanceID)
	if err != nil || len(authConfigs) == 0 {
		return nil, false
	}

	authConfig := authConfigs[0]

	// 解析kubeconfig内容（可能是JSON格式）
	var kubeconfigContent string
	if strings.HasPrefix(authConfig.ConfigValue, "{") {
		// JSON格式，需要解析
		var configData map[string]interface{}
		if err := json.Unmarshal([]byte(authConfig.ConfigValue), &configData); err == nil {
			if kubeconfig, ok := configData["kubeconfigContent"].(string); ok {
				kubeconfigContent = kubeconfig
			}
		}
	} else {
		// 直接的kubeconfig内容
		kubeconfigContent = authConfig.ConfigValue
	}

	if kubeconfigContent == "" {
		return nil, false
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return nil, false
	}

	metricsClient, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		return nil, false
	}

	return metricsClient, true
}
