package model

type MenuForm struct {
	ID         int64  `json:"id"`
	ParentID   int64  `json:"parentId"`
	Name       string `json:"name" binding:"required"`
	Type       int    `json:"type" binding:"required,oneof=1 2 3 4"`
	RouteName  string `json:"routeName"`
	RoutePath  string `json:"routePath"`
	Component  string `json:"component"`
	Perm       string `json:"perm"`
	AlwaysShow int    `json:"alwaysShow"`
	KeepAlive  int    `json:"keepAlive"`
	Visible    int    `json:"visible" binding:"oneof=0 1"`
	Sort       int    `json:"sort"`
	Icon       string `json:"icon"`
	Redirect   string `json:"redirect"`
	Params     string `json:"params"`
}
