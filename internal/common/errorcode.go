package common

// 统一错误码信息

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	return e.Msg
}

var (
	// =======================  系统相关 ========================

	UNAUTHORIZED = NewErrorCode(401, "未授权")
	BadRequest   = NewErrorCode(400, "请求错误")
	ServerError  = NewErrorCode(500, "服务器错误")
	Forbidden    = NewErrorCode(403, "无权限访问")
	// =======================  用户相关 ========================

	NotLogin             = NewErrorCode(10001, "未登录")
	UserNotExist         = NewErrorCode(10002, "用户不存在")
	UserExist            = NewErrorCode(10003, "用户已存在")
	UserPasswordError    = NewErrorCode(10004, "用户名不存在或密码错误")
	CaptchaError         = NewErrorCode(10005, "验证码错误")
	GenerateCaptchaError = NewErrorCode(10006, "生成验证码错误")
	CaptchaNotExist      = NewErrorCode(10007, "验证码不存在")
	// =======================  主机相关 ========================
	HostNotExist    = NewErrorCode(20001, "主机不存在")
	HostUnreachable = NewErrorCode(20002, "主机不可达")
	HostIdsEmpty    = NewErrorCode(20003, "主机ID不能为空")

	// =======================  菜单相关 ========================
	MenuNotExist = NewErrorCode(30001, "菜单不存在")
	MenuExist    = NewErrorCode(30002, "菜单已存在")
	// 校验菜单名称失败
	MenuNameCheckFailed = NewErrorCode(30003, "校验菜单名称失败")
	// 菜单名称在该父菜单下已存在
	MenuNameExist = NewErrorCode(30004, "菜单名称在该父菜单下已存在")
	// 该菜单下存在子菜单，不能删除
	MenuHasChildren = NewErrorCode(30005, "该菜单下存在子菜单，不能删除")

	//  =======================  job 相关 ========================
	// 获取脚本工厂失败
	ScriptFactoryNotExist = NewErrorCode(40001, "获取脚本工厂失败")
	ScriptNotExist        = NewErrorCode(40002, "脚本不存在")
	TaskNotExist          = NewErrorCode(40003, "任务不存在")

	// ========================  k8s相关 ========================

	K8sRequireTokenAndApiServer = NewErrorCode(50001, "必须填写token和api-server地址")
	K8sRequireKubeConfig        = NewErrorCode(50002, "必须填写kube-config")
)

// NewErrorCode 创建错误码，方便后续业务调用
func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{
		Code: code,
		Msg:  msg,
	}
}
