package model

type RolePageVO struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Sort       int    `json:"sort"`
	Status     int    `json:"status"`
	DataScope  int    `json:"dataScope"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}
