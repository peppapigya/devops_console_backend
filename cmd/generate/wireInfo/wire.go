//go:build wireinject
// +build wireinject

package wireInfo

import (
	"devops-console-backend/internal/controllers/system"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/redis"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
)
import "github.com/google/wire"

func InitializeLoginController() *system.LoginController {
	wire.Build(configs.NewDB, database.InitRedis, redis.NewClient, mapper.NewUserMapper, system.NewLoginController)
	return &system.LoginController{}
}
