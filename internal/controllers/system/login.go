package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/models/request/system"
	"devops-console-backend/pkg/utils"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type LoginController struct {
	userMapper *mapper.UserMapper
	redisCli   *redis.Client
}

func NewLoginController(userMapper *mapper.UserMapper, redisCli *redis.Client) *LoginController {
	return &LoginController{
		userMapper: userMapper,
		redisCli:   redisCli,
	}
}

func (l *LoginController) Login(c *gin.Context) {
	var loginRequest system.LoginRequest
	parseRequestBody(c, &loginRequest)
	// 1. 查询用户信息
}

// 登出
func (l *LoginController) Logout(c *gin.Context) {
}

func parseRequestBody(c *gin.Context, body interface{}) {
	if ok := utils.BindAndValidate(c, body); !ok {
		log.Printf("参数解析失败或验证失败\n")
		common.Fail(c, common.BadRequest)
		c.Abort()
	}
}
