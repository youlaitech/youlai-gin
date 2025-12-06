package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// 用于生成 bcrypt 密码的工具
func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run password_gen.go <密码>")
		fmt.Println("示例: go run password_gen.go 123456")
		os.Exit(1)
	}

	password := os.Args[1]
	
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("生成密码失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("加密密码: %s\n", string(hash))
	fmt.Println("\nSQL 示例:")
	fmt.Printf("INSERT INTO sys_user (username, nickname, password, mobile, status, dept_id)\n")
	fmt.Printf("VALUES ('testuser', '测试用户', '%s', '13800138000', 1, 1);\n", string(hash))
}
