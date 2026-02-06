package errs

import (
	"errors"
	"net/http"

	"youlai-gin/pkg/constant"
)

// AppError 系统通用错误类型
// Code：业务码
// Msg：提示文案
// HTTPStatus：HTTP 状态码
// Err：底层错误
type AppError struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Msg
}

// Wrap 允许 AppError 包装底层 error
func Wrap(base *AppError, err error) *AppError {
	return &AppError{
		Code:       base.Code,
		Msg:        base.Msg,
		HTTPStatus: base.HTTPStatus,
		Err:        err,
	}
}

// BadRequest 用户端错误（A0400）
// msg 为空时使用默认值
func BadRequest(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgBadRequest
	}
	return &AppError{
		Code:       constant.CodeBadRequest,
		Msg:        msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

// SystemError 系统执行出错（B0001）
// msg 为空时使用默认值
func SystemError(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgSystemError
	}
	return &AppError{
		Code:       constant.CodeSystemError,
		Msg:        msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

// New 创建自定义业务错误
func New(code, msg string, status int) *AppError {
	return &AppError{Code: code, Msg: msg, HTTPStatus: status}
}

// As 判断是否 AppError
func As(err error) (*AppError, bool) {
	var ae *AppError
	ok := errors.As(err, &ae)
	return ae, ok
}

// ========== 常用业务错误 ==========

// UserNotFound 用户不存在
func UserNotFound() *AppError {
	return &AppError{
		Code:       constant.CodeUserNotExist,
		Msg:        constant.MsgUserNotExist,
		HTTPStatus: http.StatusBadRequest,
	}
}

// TokenInvalid 令牌无效
func TokenInvalid() *AppError {
	return &AppError{
		Code:       constant.CodeAccessTokenInvalid,
		Msg:        constant.MsgAccessTokenInvalid,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NotFound 资源不存在
func NotFound(msg string) *AppError {
	if msg == "" {
		msg = "资源不存在"
	}
	return &AppError{
		Code:       constant.CodeBadRequest,
		Msg:        msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Unauthorized 未授权访问
func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgAccessUnauthorized
	}
	return &AppError{
		Code:       constant.CodeAccessUnauthorized,
		Msg:        msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}
