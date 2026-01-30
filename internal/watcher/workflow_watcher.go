package watcher

import (
	"devops-console-backend/internal/dal/model"
	"log"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	versioned "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/argoproj/argo-workflows/v3/pkg/client/informers/externalversions"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/cache" // 确保导入了这个包
)

type WorkflowWatcher struct {
	argoClient versioned.Interface
	db         *gorm.DB
}

func NewWorkflowWatcher(client versioned.Interface, db *gorm.DB) *WorkflowWatcher {
	return &WorkflowWatcher{argoClient: client, db: db}
}

func (w *WorkflowWatcher) Run(stopCh <-chan struct{}) {
	// 初始化 Factory，每 30 秒进行一次全量同步（Resync）
	factory := externalversions.NewSharedInformerFactory(w.argoClient, 0)
	informer := factory.Argoproj().V1alpha1().Workflows().Informer()

	// 修正后的类型：cache.ResourceEventHandlerFuncs
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// 当 Workflow 更新时触发（包括状态变更）
		UpdateFunc: func(oldObj, newObj interface{}) {
			wf := newObj.(*wfv1.Workflow)

			// 过滤标签
			if wf.Labels["managed-by"] != "my-devops-system" {
				return
			}

			log.Printf("检测到 Workflow 变更: %s, 当前状态: %s", wf.Name, wf.Status.Phase)

			// 更新数据库逻辑
			err := w.db.Model(&model.PipelineRun{}).
				Where("workflow_name = ?", wf.Name).
				Updates(map[string]interface{}{
					"status":     string(wf.Status.Phase),
					"start_time": wf.Status.StartedAt.Time,
					"end_time":   wf.Status.FinishedAt.Time,
				}).Error

			if err != nil {
				log.Printf("数据库更新失败: %v", err)
			}
		},
	})

	// 启动 Informer
	go factory.Start(stopCh)

	// 等待缓存同步
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		log.Fatal("同步 Argo 工作流缓存失败")
		return
	}

	log.Println("Workflow Watcher 已成功启动并同步缓存...")
	<-stopCh
}
