package service

import (
	"errors"
	"log/slog"
	"strings"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/util"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户业务逻辑。
type UserService struct {
	repo   *repository.UserRepository
	logger *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(repo *repository.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

// Register 注册用户（律师）。
func (s *UserService) Register(username, password, realName, role, licenseNo, email, phone string) (*model.User, error) {
	s.logger.Info(constants.LogUserRegisterStart, "username", username)
	username = strings.TrimSpace(username)
	if len(password) < 6 {
		return nil, util.NewAppError(constants.CodeValidationFailed, "User[username="+username+"] register: password too short")
	}
	if _, err := s.repo.FindByUsername(username); err == nil {
		return nil, util.NewAppError(constants.CodeUserExists, constants.MsgUsernameExists)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, util.Wrap(err, "User[username=%s] register check failed", username)
	}
	if role == "" {
		role = constants.RoleLawyer
	}
	if !contains(constants.UserRoleValues, role) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "User[role="+role+"] register: invalid role")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.Wrap(err, "User[username=%s] register hash failed", username)
	}
	u := &model.User{Username: username, PasswordHash: string(hash), RealName: realName,
		Role: role, LicenseNo: licenseNo, Email: email, Phone: phone}
	if err := s.repo.Create(u); err != nil {
		return nil, util.Wrap(err, "User[username=%s] register create failed", username)
	}
	s.logger.Info(constants.LogUserRegisterSuccess, "user_id", u.ID)
	return u, nil
}

// Login 登录并返回 JWT。
func (s *UserService) Login(secret string, expireHours int, username, password string) (string, *model.User, error) {
	u, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, util.NewAppError(constants.CodeInvalidCredentials, constants.MsgInvalidCredentials)
		}
		return "", nil, util.Wrap(err, "User[username=%s] login find failed", username)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		s.logger.Warn(constants.LogUserLoginFailed, "username", username)
		return "", nil, util.NewAppError(constants.CodeInvalidCredentials, constants.MsgInvalidCredentials)
	}
	token, err := util.GenerateToken(secret, util.DurationHours(expireHours), u.ID, u.Username, u.Role)
	if err != nil {
		return "", nil, util.Wrap(err, "User[username=%s] login token failed", username)
	}
	s.logger.Info(constants.LogUserLoginSuccess, "user_id", u.ID)
	return token, u, nil
}

// GetByID 查询用户。
func (s *UserService) GetByID(id uint64) (*model.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "User[id=%d] get failed", id)
	}
	return u, nil
}

// UpdateProfile 修改个人资料。
func (s *UserService) UpdateProfile(id uint64, realName, avatar, email, phone, licenseNo string) (*model.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.Wrap(err, "User[id=%d] update profile find failed", id)
	}
	if realName != "" {
		u.RealName = realName
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	if email != "" {
		u.Email = email
	}
	if phone != "" {
		u.Phone = phone
	}
	if licenseNo != "" {
		u.LicenseNo = licenseNo
	}
	if err := s.repo.Update(u); err != nil {
		return nil, util.Wrap(err, "User[id=%d] update profile save failed", id)
	}
	s.logger.Info(constants.LogUserProfileUpdate, "user_id", u.ID)
	return u, nil
}

// ListLawyers 律师列表（用于律师分配）。
func (s *UserService) ListLawyers() ([]model.User, error) {
	return s.repo.ListLawyers()
}

// List 用户列表（管理员）。
func (s *UserService) List(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.List(page, pageSize)
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
