package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"youlai-gin/pkg/apperror"
	"youlai-gin/pkg/constant"
)

// Result 统一响应结构体
// code 参考阿里错误码规范，例如：00000/A0400/B0001/C0351
// msg 为提示文案，data 为具体数据载体

type Result struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Ok 成功且携带数据
func Ok(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Result{
        Code: constant.CodeSuccess,
        Msg:  constant.MsgSuccess,
        Data: data,
    })
}

// OkMsg 成功但只返回提示信息
func OkMsg(c *gin.Context, msg string) {
    c.JSON(http.StatusOK, Result{
        Code: constant.CodeSuccess,
        Msg:  msg,
    })
}

// Error 通用错误响应，HTTP 状态码统一为 200，通过业务 Code 区分错误类型
func Error(c *gin.Context, code, msg string) {
	c.JSON(http.StatusOK, Result{
		Code: code,
		Msg:  msg,
	})
}

// FromAppError 根据 AppError 返回
func FromAppError(c *gin.Context, ae *apperror.AppError) {
	status := ae.HTTPStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}

	c.JSON(status, Result{
		Code: ae.Code,
		Msg:  ae.Msg,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, msg string) {
    if msg == "" {
        msg = constant.MsgBadRequest
    }
    Error(c, constant.CodeBadRequest, msg)
}

// SystemError 系统异常
func SystemError(c *gin.Context, msg string) {
    if msg == "" {
        msg = constant.MsgSystemError
    }
    Error(c, constant.CodeSystemError, msg)
}

// ForbiddenWrite 演示环境禁止写入时使用（示例）
func ForbiddenWrite(c *gin.Context) {
    c.JSON(http.StatusOK, Result{
        Code: constant.CodeDatabaseAccessDenied,
        Msg:  constant.MsgDatabaseAccessDenied,
    })
}
