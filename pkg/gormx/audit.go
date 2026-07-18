package gormx

import (
	"context"

	"gorm.io/gorm"
)

// operatorKey 用于在 context 中传递当前操作人 ID，审计钩子据此填充 create_by / update_by。
const operatorKey = "gormx.operator"

// WithOperator 将操作人 ID 注入 context，审计钩子据此填充审计字段。
// 用法：repo 在调用前用 gormx.WithOperator(ctx, operatorID) 包裹 context 再传入 db.WithContext。
func WithOperator(ctx context.Context, operatorID int64) context.Context {
	return context.WithValue(ctx, operatorKey, operatorID)
}

// OperatorFromContext 从 context 取出操作人 ID。
func OperatorFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(operatorKey).(int64)
	return id, ok
}

// RegisterAuditHooks 注册全局审计钩子，自动维护 create_by / update_by 审计字段。
// 仅在 context 中存在操作人时生效；缺失操作人（如代码生成内部调用）则跳过，不破坏数据。
// 仅对实际拥有该列的表生效（如 sys_menu 无审计列则自动跳过），避免写入不存在的列。
func RegisterAuditHooks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("gormx:audit_create", func(tx *gorm.DB) {
		injectAudit(tx, true)
	})
	db.Callback().Update().Before("gorm:update").Register("gormx:audit_update", func(tx *gorm.DB) {
		injectAudit(tx, false)
	})
}

func injectAudit(tx *gorm.DB, create bool) {
	id, ok := OperatorFromContext(tx.Statement.Context)
	if !ok || tx.Statement.Schema == nil {
		return
	}
	if create {
		if _, has := tx.Statement.Schema.FieldsByDBName["create_by"]; has {
			tx.Statement.SetColumn("create_by", id, false)
		}
	}
	if _, has := tx.Statement.Schema.FieldsByDBName["update_by"]; has {
		tx.Statement.SetColumn("update_by", id, false)
	}
}
