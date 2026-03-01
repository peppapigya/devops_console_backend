package system

import (
	sysCtrl "devops-console-backend/internal/controllers/system"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/pkg/configs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterSystemRouters 注册系统管理所有路由
// 支持接收外部 db 参数复用连接池，也可自动从全局配置获取
func RegisterSystemRouters(apiGroup *gin.RouterGroup, externalDB ...*gorm.DB) {
	var db *gorm.DB
	if len(externalDB) > 0 && externalDB[0] != nil {
		db = externalDB[0]
	} else {
		db = configs.NewDB()
	}

	// 初始化 Mapper
	userMapper := mapper.NewSysUserMapper(db)
	deptMapper := mapper.NewSysDeptMapper(db)
	posMapper := mapper.NewSysPositionMapper(db)
	roleMapper := mapper.NewSysRoleMapper(db)
	menuMapper := mapper.NewSysMenuMapper(db)

	// 初始化 Controller
	userCtrl := sysCtrl.NewSysUserController(userMapper, deptMapper, posMapper, menuMapper)
	deptCtrl := sysCtrl.NewSysDeptController(deptMapper)
	posCtrl := sysCtrl.NewSysPositionController(posMapper)
	roleCtrl := sysCtrl.NewSysRoleController(roleMapper, menuMapper)
	menuCtrl := sysCtrl.NewSysMenuController(menuMapper)

	// 注册登录路由（无需认证）
	RegisterLoginRoutes(apiGroup)

	sysGroup := apiGroup.Group("/system")
	{
		// ===== 当前登录用户相关（需 JWT） =====
		authGroup := sysGroup.Group("/auth")
		{
			authGroup.GET("/info", userCtrl.GetCurrentUserInfo) // 当前用户信息+权限菜单
			authGroup.PUT("/profile", userCtrl.UpdateProfile)   // 修改个人信息
			authGroup.PUT("/password", userCtrl.ChangePassword) // 修改密码
		}

		// ===== 用户管理 =====
		userGroup := sysGroup.Group("/users")
		{
			userGroup.GET("", userCtrl.List)
			userGroup.GET("/:id", userCtrl.GetByID)
			userGroup.POST("", userCtrl.Create)
			userGroup.PUT("/:id", userCtrl.Update)
			userGroup.DELETE("/:id", userCtrl.Delete)
			userGroup.PUT("/:id/status", userCtrl.UpdateStatus)
			userGroup.PUT("/:id/reset-password", userCtrl.ResetPassword)
		}

		// ===== 部门管理 =====
		deptGroup := sysGroup.Group("/depts")
		{
			deptGroup.GET("", deptCtrl.List)
			deptGroup.POST("", deptCtrl.Create)
			deptGroup.PUT("/:id", deptCtrl.Update)
			deptGroup.DELETE("/:id", deptCtrl.Delete)
		}

		// ===== 岗位管理 =====
		posGroup := sysGroup.Group("/positions")
		{
			posGroup.GET("", posCtrl.List)
			posGroup.GET("/all", posCtrl.ListAll)
			posGroup.POST("", posCtrl.Create)
			posGroup.PUT("/:id", posCtrl.Update)
			posGroup.DELETE("/:id", posCtrl.Delete)
		}

		// ===== 角色管理 =====
		roleGroup := sysGroup.Group("/roles")
		{
			roleGroup.GET("", roleCtrl.List)
			roleGroup.GET("/all", roleCtrl.ListAll)
			roleGroup.POST("", roleCtrl.Create)
			roleGroup.PUT("/:id", roleCtrl.Update)
			roleGroup.DELETE("/:id", roleCtrl.Delete)
			roleGroup.GET("/:id/menus", roleCtrl.GetMenuIDs)
			roleGroup.PUT("/:id/menus", roleCtrl.AssignMenus)
		}

		// ===== 菜单管理 =====
		menuGroup := sysGroup.Group("/menus")
		{
			menuGroup.GET("", menuCtrl.List)
			menuGroup.POST("", menuCtrl.Create)
			menuGroup.PUT("/:id", menuCtrl.Update)
			menuGroup.DELETE("/:id", menuCtrl.Delete)
		}
	}
}
