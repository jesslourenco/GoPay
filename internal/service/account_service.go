package service

import (
	"context"
	"errors"

	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
)

type AccountService interface {
	FindAll(ctx context.Context) ([]models.Account, error)
	FindOne(ctx context.Context, id string) (models.Account, error)
	Create(ctx context.Context, name string, lastname string) (string, error)
}

var _ AccountService = (*accountServiceImpl)(nil)

type accountServiceImpl struct {
	accountRepo repository.AccountRepo
}

func NewAccountService(accountRepo repository.AccountRepo) *accountServiceImpl {
	return &accountServiceImpl{
		accountRepo: accountRepo,
	}
}

func (r *accountServiceImpl) FindAll(ctx context.Context) ([]models.Account, error) {
	return []models.Account{}, errors.New("not yet implemented")
}

func (r *accountServiceImpl) FindOne(ctx context.Context, id string) (models.Account, error) {
	return models.Account{}, errors.New("not yet implemented")
}

func (r *accountServiceImpl) Create(ctx context.Context, name string, lastname string) (string, error) {
	return "", errors.New("not yet implemented")
}
