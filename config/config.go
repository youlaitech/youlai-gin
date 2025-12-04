package config

import "fmt"

// 简单示例，实际可从 .env 或配置文件加载
var (
	DBUser     = "youlai"
	DBPassword = "123456"
	DBHost     = "www.youlai.tech"
	DBPort     = "3306"
	DBName     = "youlai_boot"
)

func Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DBUser, DBPassword, DBHost, DBPort, DBName,
	)
}
