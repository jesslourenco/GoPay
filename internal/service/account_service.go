package service

import (
	"context"

	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
)

type AccountService interface {
	GetAllAccounts(ctx context.Context) ([]models.Account, error)
	GetAccount(ctx context.Context, id string) (models.Account, error)
	CreateAccount(ctx context.Context, name string, lastname string) (string, error)
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

func (r *accountServiceImpl) GetAllAccounts(ctx context.Context) ([]models.Account, error) {
	return r.accountRepo.FindAll(ctx)
}

func (r *accountServiceImpl) GetAccount(ctx context.Context, id string) (models.Account, error) {
	return r.accountRepo.FindOne(ctx, id)
}

func (r *accountServiceImpl) CreateAccount(ctx context.Context, name string, lastname string) (string, error) {
	return r.accountRepo.Create(ctx, name, lastname)
}
