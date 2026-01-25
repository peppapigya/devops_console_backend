package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/redis"
	"devops-console-backend/internal/dal/request/system"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/jwt"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	userMapper *mapper.UserMapper
	redisCli   *redis.RedisClient
}

func NewLoginController(userMapper *mapper.UserMapper, redisCli *redis.RedisClient) *LoginController {
	return &LoginController{
		userMapper: userMapper,
		redisCli:   redisCli,
	}
}

func (l *LoginController) Login(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	var loginRequest system.LoginRequest
	parseRequestBody(c, &loginRequest)
	// 1. 查询用户信息
	user, err := l.userMapper.GetUserByUsername(loginRequest.Username)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	if user == nil || user.Password != loginRequest.Password {
		helper.Fail(common.UserPasswordError)
		return
	}
	// 2. 生成 token 信息
	accessToken, refreshToken, err := getAccessTokenAndRefreshToken(c, user)
	if err != nil {
		common.Fail(c, common.ServerError)
		return
	}
	// 3. 封装返回信息
	loginResponse := response.LoginResponse{
		Id:           int64(user.ID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     user.Username,
	}
	// 4. 将token信息设置到redis中
	_ = l.redisCli.SetWithExpiration(c, fmt.Sprintf("%v:%v", common.LoginAccessPrefix, user.ID),
		accessToken, time.Duration(common.GetGlobalConfig().Jwt.ExpireTime))
	_ = l.redisCli.SetWithExpiration(c, fmt.Sprintf("%v:%v", common.LoginAccessPrefix, user.ID),
		refreshToken, time.Duration(common.GetGlobalConfig().Jwt.RefreshExpireTime))

	// 5. 将token存储到数据库中
	err = l.InsertUserToken(user, accessToken, refreshToken)
	if err != nil {
		log.Printf("插入数据失败：%v", err.Error())
		helper.Fail(common.ServerError)
		return
	}
	helper.SuccessWithData("登录成功", "data", loginResponse)
}

// InsertUserToken 插入用户token
func (l *LoginController) InsertUserToken(user *model.SystemUser, accessToken, refreshToken string) error {
	systemUserToken := &model.SystemUserToken{
		UserID:       int64(user.ID),
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    time.Now().Add(time.Duration(common.GetGlobalConfig().Jwt.RefreshExpireTime)),
	}
	return l.userMapper.InsertSystemUserToken(systemUserToken)
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

func getAccessTokenAndRefreshToken(c *gin.Context, userDO *model.SystemUser) (string, string, error) {
	accessToken, err := jwt.GenerateJwtToken(int64(userDO.ID), userDO.Username, []string{})
	if err != nil {
		log.Printf("jwt 生成失败: %v", err)
		common.Fail(c, common.ServerError)
		return "", "", err
	}
	refreshToken, err := jwt.GenerateRefreshToken(int64(userDO.ID))
	if err != nil {
		log.Printf("refresh token 生成失败: %v", err)
		common.Fail(c, common.ServerError)
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
