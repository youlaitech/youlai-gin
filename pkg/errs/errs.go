package errs

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"youlai-gin/pkg/constant"
)

// AppError 统一业务错误类型，类似 Java 的 BusinessException
// Code：业务码（参考阿里巴巴开发手册）
// Msg：用户提示信息
// HTTPStatus：HTTP 状态码
// Err：底层错误（不返回给客户端，仅用于日志）
// Stack：调用堆栈（开发环境可记录）
type AppError struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
	Stack      string `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Msg)
}

// Unwrap 支持 errors.Is/As 链式调用
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithStack 添加调用堆栈
func (e *AppError) WithStack() *AppError {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	e.Stack = string(debugStack(pcs[:n]))
	return e
}

// WithErr 包装底层错误
func (e *AppError) WithErr(err error) *AppError {
	e.Err = err
	return e
}

func debugStack(pcs []uintptr) []byte {
	frames := runtime.CallersFrames(pcs)
	var buf []byte
	for {
		frame, more := frames.Next()
		buf = append(buf, fmt.Sprintf("\n\t%s:%d %s()", frame.File, frame.Line, frame.Function)...)
		if !more {
			break
		}
	}
	return buf
}

// New 创建自定义业务错误
func New(code, msg string, status int) *AppError {
	return &AppError{Code: code, Msg: msg, HTTPStatus: status}
}

// Wrap 包装底层错误到已有 AppError
func Wrap(base *AppError, err error) *AppError {
	return &AppError{
		Code:       base.Code,
		Msg:        base.Msg,
		HTTPStatus: base.HTTPStatus,
		Err:        err,
	}
}

// As 判断是否 AppError
func As(err error) (*AppError, bool) {
	var ae *AppError
	ok := errors.As(err, &ae)
	return ae, ok
}

// Is 判断错误码是否匹配
func Is(err error, code string) bool {
	if ae, ok := As(err); ok {
		return ae.Code == code
	}
	return false
}

// BadRequest 用户请求参数错误（A0400）
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

// InvalidParam 参数校验失败（A0402）
func InvalidParam(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgInvalidUserInput
	}
	return &AppError{
		Code:       constant.CodeInvalidUserInput,
		Msg:        msg,
		HTTPStatus: http.StatusBadRequest,
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
		HTTPStatus: http.StatusNotFound,
	}
}

// UserNotFound 用户不存在（A0201）
func UserNotFound() *AppError {
	return &AppError{
		Code:       constant.CodeUserNotExist,
		Msg:        constant.MsgUserNotExist,
		HTTPStatus: http.StatusBadRequest,
	}
}

// UserPasswordError 用户名或密码错误（A0210）
func UserPasswordError() *AppError {
	return &AppError{
		Code:       constant.CodeUserPasswordError,
		Msg:        constant.MsgUserPasswordError,
		HTTPStatus: http.StatusBadRequest,
	}
}

// TokenInvalid 访问令牌无效（A0230）
func TokenInvalid() *AppError {
	return &AppError{
		Code:       constant.CodeAccessTokenInvalid,
		Msg:        constant.MsgAccessTokenInvalid,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// RefreshTokenInvalid 刷新令牌无效（A0231）
func RefreshTokenInvalid() *AppError {
	return &AppError{
		Code:       constant.CodeRefreshTokenInvalid,
		Msg:        constant.MsgRefreshTokenInvalid,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// Unauthorized 访问未授权（A0301）
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

// Forbidden 无权限访问
func Forbidden(msg string) *AppError {
	if msg == "" {
		msg = "无权限访问"
	}
	return &AppError{
		Code:       constant.CodeAccessUnauthorized,
		Msg:        msg,
		HTTPStatus: http.StatusForbidden,
	}
}

// SystemError 系统执行出错（B0001）
func SystemError(msg string) *AppError {
	if msg == "" {
		msg = constant.MsgSystemError
	}
	return &AppError{
		Code:       constant.CodeSystemError,
		Msg:        msg,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ServiceUnavailable 服务不可用
func ServiceUnavailable(msg string) *AppError {
	if msg == "" {
		msg = "服务暂时不可用，请稍后重试"
	}
	return &AppError{
		Code:       constant.CodeSystemError,
		Msg:        msg,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// DatabaseAccessDenied 数据库访问被拒绝（演示环境）
func DatabaseAccessDenied() *AppError {
	return &AppError{
		Code:       constant.CodeDatabaseAccessDenied,
		Msg:        constant.MsgDatabaseAccessDenied,
		HTTPStatus: http.StatusForbidden,
	}
}

// ThirdPartyError 第三方服务错误
func ThirdPartyError(msg string) *AppError {
	if msg == "" {
		msg = "第三方服务异常"
	}
	return &AppError{
		Code:       "C0001",
		Msg:        msg,
		HTTPStatus: http.StatusBadGateway,
	}
}
