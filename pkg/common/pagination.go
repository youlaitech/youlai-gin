package common

// BaseQuery 分页查询参数
type BaseQuery struct {
	PageNum  int `form:"pageNum" binding:"min=1" default:"1"`
	PageSize int `form:"pageSize" binding:"min=1,max=100" default:"10"`
}

// GetOffset 计算偏移量（用于数据库分页）
func (p *BaseQuery) GetOffset() int {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	return (p.PageNum - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *BaseQuery) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return p.PageSize
}

// GetPage 获取页码（用于新分页函数）
func (p *BaseQuery) GetPage() int {
	return p.PageNum
}

// GetPageSize 获取每页大小（用于新分页函数）
func (p *BaseQuery) GetPageSize() int {
	return p.PageSize
}

type PageMeta struct {
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

type PagedData struct {
	Data interface{} `json:"data"`
	Page PageMeta    `json:"page"`
}

func NewPageMeta(pageNum int, pageSize int, total int64) PageMeta {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return PageMeta{PageNum: pageNum, PageSize: pageSize, Total: total}
}
