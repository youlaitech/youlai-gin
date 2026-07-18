// Package gormx 提供 GORM 常用扩展，重点解决官方 Updates(struct) 默认跳过零值字段的问题。
package gormx

import (
	"reflect"
	"strings"
)

// BuildPatchMap 将表单结构体转换为 GORM Updates 可直接使用的「列名 → 值」映射，
// 实现语义正确的部分更新（PATCH 风格）。
//
// 零值处理语义：
//   - 指针字段：nil 表示前端未传该字段 → 跳过（不更新）；非 nil 表示显式传值，
//     即便是 0 / "" / false 也会写入（GORM Updates(struct) 默认会跳过这类零值）。
//   - 值类型字段：直接写入其当前值（含零值）。仅用于前端必传字段，不可用于「零值合法」的可选字段。
//
// 列名解析优先级：gorm tag 的 column > json tag 的 snake_case > 字段名 snake_case。
// 跳过规则：字段名为 ID（主键）、json 标签为 "-"，或 gorm 标签为 "-" 的字段。
// 关系/计算字段（如 roleIds、targetUsers）请在表单上加 gorm:"-" 标签排除，由 service 层单独处理。
func BuildPatchMap(form any) map[string]any {
	m := map[string]any{}
	if form == nil {
		return m
	}
	v := reflect.ValueOf(form)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return m
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return m
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// 展开匿名嵌入字段，把其字段也纳入映射
		if f.Anonymous {
			for k, val := range BuildPatchMap(v.Field(i).Interface()) {
				m[k] = val
			}
			continue
		}
		jsonTag := f.Tag.Get("json")
		gormTag := f.Tag.Get("gorm")
		if gormTag == "-" || jsonTag == "-" {
			continue
		}
		// 主键不参与 SET（由 WHERE id = ? 定位）
		if f.Name == "ID" {
			continue
		}
		col := columnName(f, gormTag, jsonTag)
		if col == "" {
			continue
		}
		fv := v.Field(i)
		// 指针字段：nil 跳过，非 nil 解引用后写入（保留零值）
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			m[col] = fv.Elem().Interface()
			continue
		}
		// 值类型字段：原样写入
		m[col] = fv.Interface()
	}
	return m
}

// columnName 解析字段对应的数据库列名：
// gorm column 标签优先；其次取 json 标签首段做 snake_case；最后用字段名 snake_case 兜底。
func columnName(f reflect.StructField, gormTag, jsonTag string) string {
	for _, part := range strings.Split(gormTag, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 && kv[0] == "column" && kv[1] != "" {
			return kv[1]
		}
	}
	if jsonTag != "" {
		if name := strings.Split(jsonTag, ",")[0]; name != "" {
			return toSnakeCase(name)
		}
	}
	return toSnakeCase(f.Name)
}

// toSnakeCase 将 CamelCase 转为 snake_case，如 UserName → user_name。
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteByte(byte(r) + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
