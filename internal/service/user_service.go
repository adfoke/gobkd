package service

import (
	"context"

	authjwt "github.com/go-pkgz/auth/v2/token"

	"gobkd/internal/model"
	"gobkd/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SyncCurrentUser(ctx context.Context, authUser authjwt.User) (model.User, error) {
	return s.repo.UpsertByExternalID(ctx, model.User{
		ExternalID: authUser.ID,
		Name:       authUser.Name,
		Email:      authUser.Email,
	})
}
