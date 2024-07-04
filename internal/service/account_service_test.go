package service

import (
	"testing"

	"github.com/gopay/internal/repository"
)

func TestAccountnService_CreateAccount(t *testing.T) {
}

type accountServiceDependencies struct {
	transRepoMock *repository.MockTransactionRepo
	accRepoMock   *repository.MockAccountRepo
}

func setupAccountService(t *testing.T) (*accountServiceImpl, accountServiceDependencies) {
	deps := accountServiceDependencies{
		accRepoMock: repository.NewMockAccountRepo(t),
	}

	return NewAccountService(deps.accRepoMock), deps
}
