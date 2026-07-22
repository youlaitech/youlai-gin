package model

import (
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// MenuForm 菜单表单
type MenuForm struct {
	ID         types.BigInt      `json:"id"`                                  // 主键
	ParentID   types.BigInt      `json:"parentId"`                           // 父级ID
	Name       string            `json:"name" binding:"required"`            // 名称
	Type       string            `json:"type" binding:"required,oneof=C M B E"` // 类型(C目录/M菜单/E外链/B按钮)
	RouteName  string            `json:"routeName"`                          // 路由名称
	RoutePath  string            `json:"routePath"`                          // 路由路径
	Component  string            `json:"component"`                         // 前端组件
	ExternalURL string           `json:"externalUrl"`                       // 外链地址
	Perm       string            `json:"perm"`                              // 权限标识
	AlwaysShow int               `json:"alwaysShow"`                        // 始终显示(0/1)
	KeepAlive  int               `json:"keepAlive"`                         // 页面缓存(0/1)
	Visible    int               `json:"visible" binding:"oneof=0 1"`       // 是否可见(0/1)
	Sort       int               `json:"sort"`                              // 排序
	Icon       string            `json:"icon"`                              // 图标
	Redirect   string            `json:"redirect"`                          // 重定向
	Params     []baseModel.KeyValue `json:"params"`                         // 路由参数
}
