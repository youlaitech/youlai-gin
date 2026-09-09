package enums

// LogModule 日志模块枚举
type LogModule int

const (
	LogModuleLogin   LogModule = 1
	LogModuleUser    LogModule = 2
	LogModuleRole    LogModule = 3
	LogModuleDept    LogModule = 4
	LogModuleMenu    LogModule = 5
	LogModuleDict    LogModule = 6
	LogModuleConfig  LogModule = 7
	LogModuleFile    LogModule = 8
	LogModuleNotice  LogModule = 9
	LogModuleLog     LogModule = 10
	LogModuleCodegen LogModule = 11
	LogModuleOther   LogModule = 99
)

// LogModuleDesc 模块描述
var LogModuleDesc = map[LogModule]string{
	LogModuleLogin:   "登录",
	LogModuleUser:    "用户管理",
	LogModuleRole:    "角色管理",
	LogModuleDept:    "部门管理",
	LogModuleMenu:    "菜单管理",
	LogModuleDict:    "字典管理",
	LogModuleConfig:  "系统配置",
	LogModuleFile:    "文件管理",
	LogModuleNotice:  "通知公告",
	LogModuleLog:     "日志管理",
	LogModuleCodegen: "代码生成",
	LogModuleOther:   "其他",
}

// String 实现 fmt.Stringer 接口，返回中文描述
func (m LogModule) String() string {
	if desc, ok := LogModuleDesc[m]; ok {
		return desc
	}
	return "其他"
}

// GetLabel 从数值获取描述，兼容 int 类型参数
func GetLogModuleLabel(val int) string {
	return LogModule(val).String()
}
