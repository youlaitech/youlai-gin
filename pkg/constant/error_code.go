package constant

// 统一响应码定义，参考阿里巴巴开发手册
// 00000 正常
// A**** 用户端错误
// B**** 系统执行出错
// C**** 调用第三方服务出错

const (
    // 通用成功
    CodeSuccess = "00000"
    MsgSuccess  = "一切ok"

    // 用户请求参数错误
    CodeBadRequest = "A0400"
    MsgBadRequest  = "请求参数错误"

    // 系统执行出错
    CodeSystemError = "B0001"
    MsgSystemError  = "系统执行出错"

    // 数据库写入在演示环境禁用（示例）
    CodeDatabaseAccessDenied = "C0351"
    MsgDatabaseAccessDenied  = "演示环境已禁用数据库写入功能，请本地部署修改数据库链接或开启 Mock 模式进行体验"
)
