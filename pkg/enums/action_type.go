package enums

// ActionType 操作类型枚举
type ActionType int

const (
	ActionTypeLogin          ActionType = 1
	ActionTypeLogout         ActionType = 2
	ActionTypeInsert         ActionType = 3
	ActionTypeUpdate         ActionType = 4
	ActionTypeDelete         ActionType = 5
	ActionTypeGrant          ActionType = 6
	ActionTypeExport         ActionType = 7
	ActionTypeImport         ActionType = 8
	ActionTypeUpload         ActionType = 9
	ActionTypeDownload       ActionType = 10
	ActionTypeChangePassword ActionType = 11
	ActionTypeResetPassword  ActionType = 12
	ActionTypeEnable         ActionType = 13
	ActionTypeDisable        ActionType = 14
	ActionTypeList           ActionType = 15
	ActionTypeOther          ActionType = 99
)

// ActionTypeDesc 操作类型描述
var ActionTypeDesc = map[ActionType]string{
	ActionTypeLogin:          "登录",
	ActionTypeLogout:         "登出",
	ActionTypeInsert:         "新增",
	ActionTypeUpdate:         "修改",
	ActionTypeDelete:         "删除",
	ActionTypeGrant:          "授权",
	ActionTypeExport:         "导出",
	ActionTypeImport:         "导入",
	ActionTypeUpload:         "上传",
	ActionTypeDownload:       "下载",
	ActionTypeChangePassword: "修改密码",
	ActionTypeResetPassword:  "重置密码",
	ActionTypeEnable:         "启用",
	ActionTypeDisable:        "禁用",
	ActionTypeList:           "查询列表",
	ActionTypeOther:          "其他",
}

// String 实现 fmt.Stringer 接口，返回中文描述
func (a ActionType) String() string {
	if desc, ok := ActionTypeDesc[a]; ok {
		return desc
	}
	return "其他"
}

// GetActionTypeLabel 从数值获取描述，兼容 int 类型参数
func GetActionTypeLabel(val int) string {
	return ActionType(val).String()
}
