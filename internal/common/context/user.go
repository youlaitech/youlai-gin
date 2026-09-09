package context

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/common/auth"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/gormx"
)

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int64, error) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		return 0, errs.Unauthorized("未登录或登录已过期")
	}
	return user.UserID, nil
}

// GetCurrentUser 从上下文获取当前用户详情
func GetCurrentUser(c *gin.Context) (*auth.UserDetails, error) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		return nil, errs.Unauthorized("未登录或登录已过期")
	}
	return user, nil
}

// OperatorCtx 从 gin 上下文取出当前操作人 ID 并注入 context，供 GORM 审计钩子填充
// create_by / update_by；取不到操作人时退化为原始 gin 上下文（审计钩子会自动跳过）。
func OperatorCtx(c *gin.Context) context.Context {
	ctx := context.Context(c)
	if operatorID, err := GetCurrentUserID(c); err == nil {
		ctx = gormx.WithOperator(c, operatorID)
	}
	return ctx
}

// ParsePathParam 从路径参数中解析 int64 ID
func ParsePathParam(c *gin.Context, param string, resourceName string) (int64, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errs.BadRequest(fmt.Sprintf("无效的%sID", resourceName))
	}
	return id, nil
}

// ParseIntList 解析逗号分隔的ID列表
func ParseIntList(idsStr string, resourceName string) ([]int64, error) {
	if idsStr == "" {
		return nil, errs.BadRequest(fmt.Sprintf("%sID不能为空", resourceName))
	}

	idStrArr := strings.Split(idsStr, ",")
	ids := make([]int64, 0, len(idStrArr))
	for _, idStr := range idStrArr {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, errs.BadRequest(fmt.Sprintf("无效的%sID", resourceName))
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errs.BadRequest(fmt.Sprintf("无效的%sID", resourceName))
	}
	return ids, nil
}
