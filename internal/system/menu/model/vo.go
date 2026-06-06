package model

import "youlai-gin/pkg/types"

// MenuVO 菜单视图对象
type MenuVO struct {
	ID         types.BigInt `json:"id"` // 主键
	ParentID   types.BigInt `json:"parentId"` // 父级ID
	Name       string       `json:"name"` // 名称
	Type       string       `json:"type"` // 类型(C目录/M菜单/B按钮)
	RouteName  string       `json:"routeName"` // 路由名称
	RoutePath  string       `json:"routePath"` // 路由路径
	Component  string       `json:"component"` // 前端组件
	Perm       string       `json:"perm"` // 权限标识
	AlwaysShow int          `json:"alwaysShow"` // 始终显示
	KeepAlive  int          `json:"keepAlive"` // 页面缓存
	Visible    int          `json:"visible"` // 是否可见
	Sort       int          `json:"sort"` // 排序
	Icon       string       `json:"icon"` // 图标
	Redirect   string       `json:"redirect"` // 重定向
	CreateTime types.LocalTime `json:"createTime"` // 创建时间
	UpdateTime types.LocalTime `json:"updateTime"` // 更新时间
	Children   []*MenuVO    `json:"children,omitempty"` // 子节点
}

// RouteVO 路由视图对象（前端路由配置）
type RouteVO struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"` // 名称
	Component string      `json:"component"` // 前端组件
	Redirect  string      `json:"redirect,omitempty"` // 重定向
	Meta      *RouteMeta  `json:"meta,omitempty"`
	Children  []*RouteVO  `json:"children,omitempty"` // 子节点
}

// RouteMeta 路由元信息
type RouteMeta struct {
	Title      string `json:"title"`
	Icon       string `json:"icon,omitempty"` // 图标
	Hidden     bool   `json:"hidden,omitempty"`
	AlwaysShow bool   `json:"alwaysShow,omitempty"` // 始终显示
	KeepAlive  bool   `json:"keepAlive,omitempty"` // 页面缓存
	Params     map[string]any `json:"params,omitempty"` // 路由参数
}
