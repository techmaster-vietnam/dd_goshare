package services

import (
	"context"

	"github.com/techmaster-vietnam/dd_goshare/pkg/repositories"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// tableName: "customers", "employees"
func (s *UserService) GetByUserID(ctx context.Context, tableName string, id string) (*repositories.User, error) {
	return s.repo.GetByUserID(ctx, tableName, id)
}

func (s *UserService) GetByUsername(ctx context.Context, tableName, username string) (*repositories.User, error) {
	return s.repo.GetByUsername(ctx, tableName, username)
}