package service

import (
	"io"

	"youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/excel"
)

// UserService 用户服务接口
// 定义接口的好处：
// 1. 便于单元测试（可以 Mock）
// 2. 解耦具体实现
// 3. 符合 Go 的接口设计理念
type UserService interface {
	// 用户查询
	GetUserPage(query *model.UserPageQuery) (*common.PageResult, error)
	GetUserForm(userId int64) (*model.UserForm, error)
	GetUserProfile(userId int64) (*model.UserProfileVO, error)
	GetUserOptions() ([]common.Option[string], error)

	// 用户管理
	SaveUser(form *model.UserForm) error
	DeleteUsers(ids string) error
	UpdateUserStatus(userId int64, status int) error
	UpdateUserProfile(userId int64, form *model.UserProfileForm) error

	// 密码管理
	ResetUserPassword(userId int64, password string) error
	ChangeUserPassword(userId int64, form *model.PasswordUpdateForm) error

	// 手机号管理
	SendMobileCode(mobile string) error
	BindOrChangeMobile(userId int64, form *model.MobileUpdateForm) error

	// 邮箱管理
	SendEmailCode(email string) error
	BindOrChangeEmail(userId int64, form *model.EmailUpdateForm) error

	// Excel 导入导出
	ExportUsersToExcel(query *model.UserPageQuery) (*excel.ExcelExporter, error)
	GenerateUserTemplate() (*excel.ExcelExporter, error)
	ImportUsersFromExcel(file io.Reader) (map[string]interface{}, error)
}

// 确保 userService 实现了 UserService 接口
// 这是 Go 的编译时检查技巧
var _ UserService = (*userService)(nil)

// userService 用户服务实现
// 注意：结构体名小写（私有），接口名大写（公开）
type userService struct {
	// 可以添加依赖注入
	// repo repository.UserRepository
}

// NewUserService 创建用户服务实例
// 这是 Go 的工厂模式，类似 Java 的 @Service
func NewUserService() UserService {
	return &userService{}
}
