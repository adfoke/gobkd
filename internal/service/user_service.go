package service

import (
	"context"
	"database/sql"

	authjwt "github.com/go-pkgz/auth/v2/token"

	"gobkd/internal/apperr"
	"gobkd/internal/model"
	"gobkd/internal/repository"
)

type UserService struct {
	repo       *repository.UserRepository
	transactor *repository.Transactor
}

func NewUserService(repo *repository.UserRepository, transactor *repository.Transactor) *UserService {
	return &UserService{
		repo:       repo,
		transactor: transactor,
	}
}

func (s *UserService) SyncCurrentUser(ctx context.Context, authUser authjwt.User) (model.User, error) {
	if authUser.ID == "" {
		return model.User{}, apperr.Unauthorized("login required")
	}

	var user model.User
	err := s.transactor.WithinTransaction(ctx, func(tx *sql.Tx) error {
		res, err := s.repo.WithTx(tx).UpsertByExternalID(ctx, model.User{
			ExternalID: authUser.ID,
			Name:       authUser.Name,
			Email:      authUser.Email,
		})
		if err != nil {
			return apperr.Internal("failed to sync current user", err)
		}
		user = res
		return nil
	})
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
