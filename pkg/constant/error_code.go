package constant

// 统一响应码定义，参考阿里巴巴开发手册
// 00000 正常
// A**** 用户端错误
// B**** 系统执行出错
// C**** 调用第三方服务出错

const (
	CodeSuccess = "00000"
	MsgSuccess  = "成功"

	CodeBadRequest = "A0400" // 通用参数错误
	MsgBadRequest  = "用户请求参数错误"

	CodeUserRegistrationError = "A0100" // 用户注册错误
	MsgUserRegistrationError  = "用户注册错误"

	CodeInvalidUserInput = "A0402" // 参数校验失败
	MsgInvalidUserInput  = "无效的用户输入"

	CodeUserNotExist = "A0201" // 用户不存在
	MsgUserNotExist  = "用户账户不存在"

	CodeUserPasswordError = "A0210" // 用户名或密码错误
	MsgUserPasswordError  = "用户名或密码错误"

	CodeAccessTokenInvalid = "A0230" // 访问令牌无效
	MsgAccessTokenInvalid  = "访问令牌无效或已过期"

	CodeRefreshTokenInvalid = "A0231" // 刷新令牌无效
	MsgRefreshTokenInvalid  = "刷新令牌无效或已过期"

	CodeAccessUnauthorized = "A0301" // 访问未授权
	MsgAccessUnauthorized  = "访问未授权"

	CodeRequestConcurrencyLimitExceeded = "A0502" // 请求并发数超出限制
	MsgRequestConcurrencyLimitExceeded  = "请求并发数超出限制"

	CodeSystemError = "B0001"
	MsgSystemError  = "系统执行出错"

	CodeDatabaseAccessDenied = "C0351" // 演示环境禁用写入
	MsgDatabaseAccessDenied  = "演示环境已禁用数据库写入功能，请本地部署修改数据库链接或开启Mock模式进行体验"
)
