package executor

// InitExecutors 初始化所有执行器
func InitExecutors() {
	factory := GetExecutorFactory()

	// 注册所有执行器
	factory.Register(NewHTTPExecutor())
	factory.Register(NewScriptExecutor())
	factory.Register(NewSQLExecutor())
	factory.Register(NewK8sExecutor())
}
