package model

import "youlai-gin/pkg/types"

type MenuForm struct {
	ID         types.BigInt `json:"id"`
	ParentID   types.BigInt `json:"parentId"`
	Name       string       `json:"name" binding:"required"`
	Type       string       `json:"type" binding:"required,oneof=C M B"`
	RouteName  string       `json:"routeName"`
	RoutePath  string       `json:"routePath"`
	Component  string       `json:"component"`
	Perm       string       `json:"perm"`
	AlwaysShow int          `json:"alwaysShow"`
	KeepAlive  int          `json:"keepAlive"`
	Visible    int          `json:"visible" binding:"oneof=0 1"`
	Sort       int          `json:"sort"`
	Icon       string       `json:"icon"`
	Redirect   string       `json:"redirect"`
	Params     string       `json:"params"`
}
