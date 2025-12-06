package model

type DeptVO struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	ParentID   int64     `json:"parentId"`
	TreePath   string    `json:"treePath,omitempty"`
	Sort       int       `json:"sort"`
	Status     int       `json:"status"`
	CreateTime string    `json:"createTime,omitempty"`
	UpdateTime string    `json:"updateTime,omitempty"`
	Children   []*DeptVO `json:"children,omitempty"`
}
