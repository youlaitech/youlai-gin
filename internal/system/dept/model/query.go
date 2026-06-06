package model

// DeptQuery 部门查询参数
type DeptQuery struct {
	Keywords string `form:"keywords"` // 关键字

	Status   *int   `form:"status"` // 状态(1启用0禁用)

}
