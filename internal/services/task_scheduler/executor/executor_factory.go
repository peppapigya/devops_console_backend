package executor

import (
	"log"
	"sync"
)

// ExecutorFactory 执行器工厂
type ExecutorFactory struct {
	executors map[string]TaskExecutor
	mu        sync.RWMutex
}

var factory *ExecutorFactory
var once sync.Once

// GetExecutorFactory 获取单例工厂
func GetExecutorFactory() *ExecutorFactory {
	once.Do(func() {
		factory = &ExecutorFactory{
			executors: make(map[string]TaskExecutor),
		}
	})
	return factory
}

// Register 注册执行器
func (f *ExecutorFactory) Register(executor TaskExecutor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.executors[executor.GetType()] = executor
}

// GetExecutor 获取执行器
func (f *ExecutorFactory) GetExecutor(taskType string) (TaskExecutor, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	executor, ok := f.executors[taskType]
	test := f.ListExecutors()
	log.Printf("支持的类型：%v", test)
	return executor, ok
}

// ListExecutors 列出所有执行器类型
func (f *ExecutorFactory) ListExecutors() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	types := make([]string, 0, len(f.executors))
	for t := range f.executors {
		types = append(types, t)
	}
	return types
}
