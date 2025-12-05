package apperror

import (
	"errors"
	"net/http"

	"youlai-gin/pkg/constant"
)

// AppError 是整个系统通用错误类型
// Code：业务码（给前端）
// Msg：提示文案（给前端）
// HTTPStatus：对应 http 状态码（给后端/HTTP 层）
// Err：底层错误（给日志/链路追踪）
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
 
// ErrBadRequest 用户端错误（A0400）
// msg 为空时使用统一文案
func ErrBadRequest(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgBadRequest
	}
	return &AppError{
		Code:       constant.CodeBadRequest,
		Msg:        msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrSystem 系统执行出错（B0001）
// msg 为空时使用统一文案
func ErrSystem(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgSystemError
	}
	return &AppError{
		Code:       constant.CodeSystemError,
		Msg:        msg,
		HTTPStatus: http.StatusInternalServerError,
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
