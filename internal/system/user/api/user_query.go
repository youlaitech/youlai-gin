package api

import "youlai-gin/pkg/common"

// UserPageQueryReq 用户分页查询参数
type UserQueryReq struct {
	common.BaseQuery
	Keywords   string  `form:"keywords"`   // 关键字(用户名/昵称/手机号)
	Status     *int    `form:"status"`     // 用户状态
	DeptID     *int64  `form:"deptId"`     // 部门ID
	CreateTime []string `form:"createTime"` // 创建时间范围
}
