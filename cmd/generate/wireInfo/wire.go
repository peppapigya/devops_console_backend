//go:build wireinject
// +build wireinject

package wireInfo

import (
	"devops-console-backend/internal/controllers/cicd"
	"devops-console-backend/internal/controllers/system"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/redis"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
	"devops-console-backend/pkg/utils/jwt"
)
import "github.com/google/wire"

func InitializeLoginController() *system.LoginController {
	wire.Build(configs.NewDB, database.InitRedis, redis.NewClient, jwt.NewBlackListManager, mapper.NewUserMapper, system.NewLoginController)
	return &system.LoginController{}
}
func InitializePipelineController() *cicd.PipelinesController {
	wire.Build(configs.NewDB, mapper.NewPipelinesMapper, cicd.NewPipelinesController)
	return &cicd.PipelinesController{}
}

func InitializePipelineRunsController() *cicd.PipelineRunController {
	wire.Build(configs.NewDB, mapper.NewPipelineRunMapper, cicd.NewPipelineRunController)
	return &cicd.PipelineRunController{}
}
func InitializeProjectsController() *cicd.ProjectsController {
	wire.Build(configs.NewDB, mapper.NewProjectMapper, cicd.NewProjectsController)
	return &cicd.ProjectsController{}
}
func InitializeArgoController() *cicd.ArgoController {
	wire.Build(configs.NewDB, mapper.NewPipelineRunMapper, mapper.NewPipelinesMapper, mapper.NewPipelineStepsMapper, cicd.NewArgoController)
	return &cicd.ArgoController{}
}
func InitializePipelineStepsController() *cicd.PipelineStepsController {
	wire.Build(configs.NewDB, mapper.NewPipelineStepsMapper, cicd.NewPipelineStepsController)
	return &cicd.PipelineStepsController{}
}
