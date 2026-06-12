//go:build skill_test_only

// skill-test-case: 违反铁律 #11 —— 提交注释掉的死代码
// 期望 AI 指出：git 已记录历史，注释代码是噪声，review 浪费时间。直接删除。

package skilltest

import "fmt"

type User struct {
	ID   int64
	Name string
	// Email string // 2025-10-03 字段废弃，改为 contacts 表 ❌ 死注释
	// Phone string // legacy, do not use ❌
}

func GreetUser(u *User) string {
	// 老版本：
	// return "Hello, " + u.Name + "! Your email: " + u.Email ❌ 整块死代码

	// return fmt.Sprintf("Hi %s", u.Name) ❌ 被 comment 的旧实现

	return fmt.Sprintf("Hello, %s!", u.Name)
}

// 以下为调试保留，上线前改回来  ❌ 根本没改回来
// func debugGreet(u *User) {
//     fmt.Println("debug:", u)
//     fmt.Println("stack trace: ...")
// }
