package model

type Menu struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID   int64  `gorm:"column:parent_id;not null" json:"parentId"`
	TreePath   string `gorm:"column:tree_path" json:"treePath"`
	Name       string `gorm:"column:name;not null" json:"name"`
	Type       int    `gorm:"column:type;not null" json:"type"`
	RouteName  string `gorm:"column:route_name" json:"routeName"`
	RoutePath  string `gorm:"column:route_path" json:"routePath"`
	Component  string `gorm:"column:component" json:"component"`
	Perm       string `gorm:"column:perm" json:"perm"`
	AlwaysShow int    `gorm:"column:always_show;default:0" json:"alwaysShow"`
	KeepAlive  int    `gorm:"column:keep_alive;default:0" json:"keepAlive"`
	Visible    int    `gorm:"column:visible;default:1" json:"visible"`
	Sort       int    `gorm:"column:sort;default:0" json:"sort"`
	Icon       string `gorm:"column:icon" json:"icon"`
	Redirect   string `gorm:"column:redirect" json:"redirect"`
	CreateTime string `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime string `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	Params     string `gorm:"column:params" json:"params"`
}

func (Menu) TableName() string {
	return "sys_menu"
}
