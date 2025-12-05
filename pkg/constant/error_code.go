package constant

// 统一响应码定义，参考阿里巴巴开发手册
// 00000 正常
// A**** 用户端错误
// B**** 系统执行出错
// C**** 调用第三方服务出错

const (
    // ========== 通用成功 ==========
    CodeSuccess = "00000"
    MsgSuccess  = "一切ok"

    // ========== A0*** 用户端错误（典型示例）==========
    CodeBadRequest = "A0400" // 通用参数错误
    MsgBadRequest  = "请求参数错误"

    CodeInvalidUserInput = "A0402" // 参数校验失败（更具体）
    MsgInvalidUserInput  = "无效的用户输入"

    CodeUserNotExist = "A0201" // 用户不存在
    MsgUserNotExist  = "用户不存在"

    CodeUsernameExists = "A0111" // 用户名已存在
    MsgUsernameExists  = "用户名已存在"

    CodeAccessTokenInvalid = "A0230" // 令牌无效
    MsgAccessTokenInvalid  = "访问令牌无效或已过期"

    // ========== B0*** 系统执行出错 ==========
    CodeSystemError = "B0001"
    MsgSystemError  = "系统执行出错"

    // ========== C0*** 第三方服务出错 ==========
    CodeDatabaseAccessDenied = "C0351" // 演示环境禁用写入（示例）
    MsgDatabaseAccessDenied  = "演示环境已禁用数据库写入功能，请本地部署修改数据库链接或开启 Mock 模式进行体验"
)
