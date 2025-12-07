package common

// PageQuery 分页查询参数
type PageQuery struct {
	PageNum  int `form:"pageNum" binding:"min=1" default:"1"`
	PageSize int `form:"pageSize" binding:"min=1,max=100" default:"10"`
}

// GetOffset 计算偏移量（用于数据库分页）
func (p *PageQuery) GetOffset() int {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	return (p.PageNum - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PageQuery) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return p.PageSize
}

// GetPage 获取页码（用于新分页函数）
func (p *PageQuery) GetPage() int {
	return p.PageNum
}

// GetPageSize 获取每页大小（用于新分页函数）
func (p *PageQuery) GetPageSize() int {
	return p.PageSize
}

// PageResult 分页响应
type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

// Option 下拉选项（泛型）
type Option[T any] struct {
	Value T      `json:"value"`
	Label string `json:"label"`
}

// BaseEntity 基础实体（所有表的公共字段）
type BaseEntity struct {
	CreateBy   *int64 `gorm:"column:create_by" json:"createBy,omitempty"`
	CreateTime string `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
	UpdateBy   *int64 `gorm:"column:update_by" json:"updateBy,omitempty"`
	UpdateTime string `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	IsDeleted  int    `gorm:"column:is_deleted;default:0" json:"-"`
}
