package model

type DeptForm struct {
	ID       int64  `json:"id"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	ParentID int64  `json:"parentId"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status" binding:"oneof=0 1"`
}
