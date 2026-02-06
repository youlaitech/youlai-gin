package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/constant"
)

// Result 统一响应结构体
// code 参考阿里错误码规范，如 00000/A0400/B0001/C0351
// msg 为提示文案，data 为数据载体

type Result struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Page interface{} `json:"page"`
}

type Page struct {
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

// Ok 成功且携带数据
func Ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{
		Code: constant.CodeSuccess,
		Msg:  constant.MsgSuccess,
		Data: data,
		Page: nil,
	})
}

func OkPage(c *gin.Context, data interface{}, pageNum int, pageSize int, total int64) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	c.JSON(http.StatusOK, Result{
		Code: constant.CodeSuccess,
		Msg:  constant.MsgSuccess,
		Data: data,
		Page: &Page{PageNum: pageNum, PageSize: pageSize, Total: total},
	})
}

func OkPaged(c *gin.Context, paged *common.PagedData) {
	if paged == nil {
		OkPage(c, nil, 1, 10, 0)
		return
	}

	c.JSON(http.StatusOK, Result{
		Code: constant.CodeSuccess,
		Msg:  constant.MsgSuccess,
		Data: paged.Data,
		Page: paged.Page,
	})
}

// OkMsg 成功但只返回提示信息
func OkMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Result{
		Code: constant.CodeSuccess,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, msg string) {
	if msg == "" {
		msg = constant.MsgBadRequest
	}
	c.JSON(http.StatusBadRequest, Result{
		Code: constant.CodeBadRequest,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// FromAppError 根据 AppError 返回
func FromAppError(c *gin.Context, ae *errs.AppError) {
	status := ae.HTTPStatus
	if status == 0 {
		status = http.StatusBadRequest
	}

	c.JSON(status, Result{
		Code: ae.Code,
		Msg:  ae.Msg,
		Data: nil,
		Page: nil,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = constant.MsgBadRequest
	}
	c.JSON(http.StatusBadRequest, Result{
		Code: constant.CodeBadRequest,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// Unauthorized 访问未授权（已登录但无权限）
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = constant.MsgAccessUnauthorized
	}
	c.JSON(http.StatusUnauthorized, Result{
		Code: constant.CodeAccessUnauthorized,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// TokenInvalid 访问令牌无效
func TokenInvalid(c *gin.Context, msg string) {
	if msg == "" {
		msg = constant.MsgAccessTokenInvalid
	}
	c.JSON(http.StatusUnauthorized, Result{
		Code: constant.CodeAccessTokenInvalid,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// SystemError 系统异常
func SystemError(c *gin.Context, msg string) {
	if msg == "" {
		msg = constant.MsgSystemError
	}
	c.JSON(http.StatusBadRequest, Result{
		Code: constant.CodeSystemError,
		Msg:  msg,
		Data: nil,
		Page: nil,
	})
}

// InternalServerError 内部服务器错误（SystemError 的别名）
func InternalServerError(c *gin.Context, msg string) {
	SystemError(c, msg)
}

// ForbiddenWrite 演示环境禁止写入
func ForbiddenWrite(c *gin.Context) {
	c.JSON(http.StatusForbidden, Result{
		Code: constant.CodeDatabaseAccessDenied,
		Msg:  constant.MsgDatabaseAccessDenied,
		Data: nil,
		Page: nil,
	})
}

// HandleError 统一错误处理
func HandleError(c *gin.Context, err error) {
	if ae, ok := err.(*errs.AppError); ok {
		FromAppError(c, ae)
	} else {
		SystemError(c, err.Error())
	}
	return
}
