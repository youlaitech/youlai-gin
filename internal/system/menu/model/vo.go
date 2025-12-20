package model

// MenuVO 菜单视图对象
type MenuVO struct {
	ID         int64      `json:"id"`
	ParentID   int64      `json:"parentId"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	RouteName  string     `json:"routeName"`
	RoutePath  string     `json:"routePath"`
	Component  string     `json:"component"`
	Perm       string     `json:"perm"`
	AlwaysShow int        `json:"alwaysShow"`
	KeepAlive  int        `json:"keepAlive"`
	Visible    int        `json:"visible"`
	Sort       int        `json:"sort"`
	Icon       string     `json:"icon"`
	Redirect   string     `json:"redirect"`
	CreateTime string     `json:"createTime"`
	UpdateTime string     `json:"updateTime"`
	Children   []*MenuVO  `json:"children,omitempty"`
}

// RouteVO 路由视图对象（前端路由配置）
type RouteVO struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Component string      `json:"component"`
	Redirect  string      `json:"redirect,omitempty"`
	Meta      *RouteMeta  `json:"meta,omitempty"`
	Children  []*RouteVO  `json:"children,omitempty"`
}

// RouteMeta 路由元信息
type RouteMeta struct {
	Title      string `json:"title"`
	Icon       string `json:"icon,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	AlwaysShow bool   `json:"alwaysShow,omitempty"`
	KeepAlive  bool   `json:"keepAlive,omitempty"`
	Params     string `json:"params,omitempty"`
}
