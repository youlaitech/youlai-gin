package service

import (
	"io"

	"youlai-gin/internal/system/user/api"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/excel"
)

// UserService 用户服务接口
type UserService interface {
	// 用户查询
	GetUserPage(query *api.UserQueryReq) (*common.PagedData, error)
	GetUserForm(userId int64) (*api.UserFormResp, error)
	GetUserProfile(userId int64) (*api.UserProfileResp, error)
	GetUserOptions() ([]common.Option[string], error)

	// 用户管理
	SaveUser(form *api.UserSaveReq) error
	DeleteUsers(ids string) error
	UpdateUserStatus(userId int64, status int) error
	UpdateUserProfile(userId int64, req *api.UserProfileUpdateReq) error

	// 密码管理
	ResetUserPassword(userId int64, password string) error
	ChangeUserPassword(userId int64, form *api.PasswordUpdateReq) error

	// 手机号管理
	SendMobileCode(mobile string) error
	BindOrChangeMobile(userId int64, form *api.MobileUpdateReq) error
	UnbindMobile(userId int64, form *api.PasswordVerifyReq) error

	// 邮箱管理
	SendEmailCode(email string) error
	BindOrChangeEmail(userId int64, form *api.EmailUpdateReq) error
	UnbindEmail(userId int64, form *api.PasswordVerifyReq) error

	// Excel 导入导出
	ExportUsersToExcel(query *api.UserQueryReq) (*excel.ExcelExporter, error)
	GenerateUserTemplate() (*excel.ExcelExporter, error)
	ImportUsersFromExcel(file io.Reader) (map[string]interface{}, error)
}

// userService 实现检查
var _ UserService = (*userService)(nil)

// userService 用户服务实现
type userService struct {
}

func (s *userService) GetUserPage(query *api.UserQueryReq) (*common.PagedData, error) {
	return GetUserPage(query)
}

func (s *userService) GetUserForm(userId int64) (*api.UserFormResp, error) {
	return GetUserForm(userId)
}

func (s *userService) GetUserProfile(userId int64) (*api.UserProfileResp, error) {
	return GetUserProfile(userId)
}

func (s *userService) GetUserOptions() ([]common.Option[string], error) {
	return GetUserOptions()
}

func (s *userService) SaveUser(form *api.UserSaveReq) error {
	return SaveUser(form)
}

func (s *userService) DeleteUsers(ids string) error {
	return DeleteUsers(ids)
}

func (s *userService) UpdateUserStatus(userId int64, status int) error {
	return UpdateUserStatus(userId, status)
}

func (s *userService) UpdateUserProfile(userId int64, req *api.UserProfileUpdateReq) error {
	return UpdateUserProfile(userId, req)
}

func (s *userService) ResetUserPassword(userId int64, password string) error {
	return ResetUserPassword(userId, password)
}

func (s *userService) ChangeUserPassword(userId int64, form *api.PasswordUpdateReq) error {
	return ChangeUserPassword(userId, form)
}

func (s *userService) SendMobileCode(mobile string) error {
	return SendMobileCode(mobile)
}

func (s *userService) BindOrChangeMobile(userId int64, form *api.MobileUpdateReq) error {
	return BindOrChangeMobile(userId, form)
}

func (s *userService) UnbindMobile(userId int64, form *api.PasswordVerifyReq) error {
	return UnbindMobile(userId, form)
}

func (s *userService) SendEmailCode(email string) error {
	return SendEmailCode(email)
}

func (s *userService) BindOrChangeEmail(userId int64, form *api.EmailUpdateReq) error {
	return BindOrChangeEmail(userId, form)
}

func (s *userService) UnbindEmail(userId int64, form *api.PasswordVerifyReq) error {
	return UnbindEmail(userId, form)
}

func (s *userService) ExportUsersToExcel(query *api.UserQueryReq) (*excel.ExcelExporter, error) {
	return ExportUsersToExcel(query)
}

func (s *userService) GenerateUserTemplate() (*excel.ExcelExporter, error) {
	return GenerateUserTemplate()
}

func (s *userService) ImportUsersFromExcel(file io.Reader) (map[string]interface{}, error) {
	return ImportUsersFromExcel(file)
}

// NewUserService 创建用户服务实例
func NewUserService() UserService {
	return &userService{}
}
