package model

import (
	"time"

	"youlai-gin/pkg/types"
)

// Log 操作日志实体（对应 sys_log 表）
type Log struct {
	ID            types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	Module        int         `gorm:"column:module" json:"module"`
	ActionType    int         `gorm:"column:action_type" json:"actionType"`
	Title         string      `gorm:"column:title;size:100" json:"title"`
	Content       string      `gorm:"column:content;type:text" json:"content"`
	OperatorID    types.BigInt `gorm:"column:operator_id" json:"operatorId"`
	OperatorName  string      `gorm:"column:operator_name;size:50" json:"operatorName"`
	RequestURI    string      `gorm:"column:request_uri;size:255" json:"requestUri"`
	RequestMethod string      `gorm:"column:request_method;size:10" json:"requestMethod"`
	IP            string      `gorm:"column:ip;size:45" json:"ip"`
	Province      string      `gorm:"column:province;size:100" json:"province"`
	City          string      `gorm:"column:city;size:100" json:"city"`
	Device        string      `gorm:"column:device;size:100" json:"device"`
	OS            string      `gorm:"column:os;size:100" json:"os"`
	Browser       string      `gorm:"column:browser;size:100" json:"browser"`
	Status        int         `gorm:"column:status" json:"status"`
	ErrorMsg      string      `gorm:"column:error_msg;size:255" json:"errorMsg"`
	ExecutionTime int         `gorm:"column:execution_time" json:"executionTime"`
	CreateTime    time.Time   `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (Log) TableName() string {
	return "sys_log"
}
