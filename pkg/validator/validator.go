package validator

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/constant"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// BindJSON 参数绑定和校验
func BindJSON(c *gin.Context, dst any) error {
	// 1. 绑定 JSON
	if err := c.ShouldBindJSON(dst); err != nil {
		return errs.Wrap(
			errs.BadRequest("JSON 格式错误"),
			err,
		)
	}

	// 2. 执行 validate 校验
	if err := validate.Struct(dst); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// 提取所有字段错误
			var messages []string
			for _, fieldError := range validationErrors {
				messages = append(messages, translateFieldError(fieldError))
			}
			msg := strings.Join(messages, "；")
			return errs.New(
				constant.CodeInvalidUserInput,
				msg,
				400,
			)
		}
		return errs.BadRequest(err.Error())
	}

	return nil
}

// BindQuery Query 参数绑定和校验
func BindQuery(c *gin.Context, dst any) error {
	// 1. 绑定 Query 参数
	if err := c.ShouldBindQuery(dst); err != nil {
		return errs.Wrap(
			errs.BadRequest("查询参数错误"),
			err,
		)
	}

	// 2. 执行 validate 校验
	if err := validate.Struct(dst); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fieldError := range validationErrors {
				messages = append(messages, translateFieldError(fieldError))
			}
			msg := strings.Join(messages, "；")
			return errs.New(
				constant.CodeInvalidUserInput,
				msg,
				400,
			)
		}
		return errs.BadRequest(err.Error())
	}

	return nil
}

// BindURI URI 参数绑定和校验
func BindURI(c *gin.Context, dst any) error {
	// 1. 绑定 URI 参数
	if err := c.ShouldBindUri(dst); err != nil {
		return errs.Wrap(
			errs.BadRequest("路径参数错误"),
			err,
		)
	}

	// 2. 执行 validate 校验
	if err := validate.Struct(dst); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fieldError := range validationErrors {
				messages = append(messages, translateFieldError(fieldError))
			}
			msg := strings.Join(messages, "；")
			return errs.New(
				constant.CodeInvalidUserInput,
				msg,
				400,
			)
		}
		return errs.BadRequest(err.Error())
	}

	return nil
}

// Validate 直接校验结构体
func Validate(dst any) error {
	if err := validate.Struct(dst); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fieldError := range validationErrors {
				messages = append(messages, translateFieldError(fieldError))
			}
			msg := strings.Join(messages, "；")
			return errs.New(
				constant.CodeInvalidUserInput,
				msg,
				400,
			)
		}
		return errs.BadRequest(err.Error())
	}
	return nil
}

// translateFieldError 转换字段错误
func translateFieldError(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "email":
		return fmt.Sprintf("%s格式不正确", field)
	case "min":
		return fmt.Sprintf("%s长度不能少于%s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s长度不能超过%s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s长度必须为%s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s必须大于%s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s必须大于等于%s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s必须小于%s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s必须小于等于%s", field, fe.Param())
	default:
		return fmt.Sprintf("%s校验失败: %s", field, tag)
	}
}
