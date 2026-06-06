package model

// MenuQuery 菜单查询参数
type MenuQuery struct {
	Keywords string `form:"keywords"` // 关键字

	Status   *int   `form:"status"` // 状态

}
