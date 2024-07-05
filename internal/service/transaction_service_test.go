package service

import (
	"context"
	"testing"
	"time"

	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
	"github.com/gopay/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestTransactionService_GetBalance(t *testing.T) {
	var (
		ctx   = context.Background()
		owner = "0001"
	)

	type args struct {
		owner string
	}

	scenarios := map[string]struct {
		given   args
		doMocks func(deps transactionServiceDependencies)
		want    models.Balance
		wantErr error
	}{
		"happy-path": {
			given: args{
				owner: owner,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    10000,
				}, nil)
			},
			want: models.Balance{
				AccountId: owner,
				Amount:    10000,
			},
			wantErr: nil,
		},
		"invalid-account": {
			given: args{
				owner: "1234",
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, "1234").Return(models.Account{}, repository.ErrAccountNotFound)
			},
			want:    models.Balance{},
			wantErr: repository.ErrAccountNotFound,
		},
	}

	for name, tcase := range scenarios {
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			service, deps := setupTransactionService(t)
			if tcase.doMocks != nil {
				tcase.doMocks(deps)
			}

			balance, err := service.GetBalance(ctx, tcase.given.owner)

			if tcase.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tcase.wantErr)
			}

			assert.Equal(t, tcase.want, balance)
		})
	}
}

func TestTransactionService_GetAllTransactions(t *testing.T) {
	now := time.Now()

	var (
		ctx   = context.Background()
		owner = "0001"
		data  = []models.Transaction{
			{
				TransactionId: "1000000",
				Owner:         owner,
				Sender:        owner,
				Receiver:      owner,
				CreatedAt:     now,
				Amount:        7000.00,
				IsConsumed:    false,
			},
			{
				TransactionId: "2000000",
				Owner:         owner,
				Sender:        owner,
				Receiver:      owner,
				CreatedAt:     now,
				Amount:        3000.00,
				IsConsumed:    false,
			},
		}
	)

	type args struct {
		owner string
	}

	scenarios := map[string]struct {
		given   args
		doMocks func(deps transactionServiceDependencies)
		want    []models.Transaction
		wantErr error
	}{
		"happy-path": {
			given: args{
				owner: owner,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("FindAll", ctx, owner).Return(data, nil)
			},
			want:    data,
			wantErr: nil,
		},
		"invalid-account": {
			given: args{
				owner: "1234",
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, "1234").Return(models.Account{}, repository.ErrAccountNotFound)
			},
			want:    []models.Transaction{},
			wantErr: repository.ErrAccountNotFound,
		},
	}

	for name, tcase := range scenarios {
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			service, deps := setupTransactionService(t)
			if tcase.doMocks != nil {
				tcase.doMocks(deps)
			}

			transactions, err := service.GetAllTransactions(ctx, tcase.given.owner)

			if tcase.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tcase.wantErr)
			}

			assert.ElementsMatch(t, tcase.want, transactions)
		})
	}
}

func TestTransactionService_Deposit(t *testing.T) {
	now := time.Now()
	setupClock(now)
	defer resetClock()

	var (
		ctx            = context.Background()
		owner          = "0001"
		amount float32 = 50.0
	)

	type args struct {
		owner  string
		amount float32
	}

	scenarios := map[string]struct {
		given   args
		doMocks func(deps transactionServiceDependencies)
		want    models.Transaction
		wantErr error
	}{
		"happy-path": {
			given: args{
				owner:  owner,
				amount: amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     amount,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)
				deps.transRepoMock.On("Create", ctx, transaction).Return(nil)
			},
			wantErr: nil,
		},
		"zero-amount": {
			given: args{
				owner:  owner,
				amount: 0.00,
			},
			wantErr: ErrInvalidAmount,
		},
		"negative-amount": {
			given: args{
				owner:  owner,
				amount: -1.00,
			},
			wantErr: ErrInvalidAmount,
		},
		"invalid-owner": {
			given: args{
				owner:  owner,
				amount: amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{}, repository.ErrAccountNotFound)
			},
			wantErr: repository.ErrAccountNotFound,
		},
	}

	for name, tcase := range scenarios {
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			service, deps := setupTransactionService(t)
			if tcase.doMocks != nil {
				tcase.doMocks(deps)
			}

			err := service.Deposit(ctx, tcase.given.owner, tcase.given.amount)

			if tcase.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tcase.wantErr)
			}
		})
	}
}

func TestTransactionService_Withdraw(t *testing.T) {
	now := time.Now()
	setupClock(now)
	utils.SetSyncGoroutine()
	defer utils.ResetGoroutine()
	defer resetClock()

	var (
		ctx            = context.Background()
		owner          = "0001"
		amount float32 = -1000.0
	)

	type args struct {
		owner  string
		amount float32
	}

	scenarios := map[string]struct {
		given   args
		doMocks func(deps transactionServiceDependencies)
		wantErr error
	}{
		"happy-path": {
			given: args{
				owner:  owner,
				amount: amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        7000.0,
					},
				}

				debitTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: true,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     amount,
				}

				transaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     7000 + amount,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    7000,
				}, nil)
				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("Create", ctx, transaction).Return(nil)
				deps.transRepoMock.On("Create", ctx, debitTransaction).Return(nil)
			},
			wantErr: nil,
		},
		"invalid-owner": {
			given: args{
				owner:  owner,
				amount: amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{}, repository.ErrAccountNotFound)
			},
			wantErr: repository.ErrAccountNotFound,
		},
		"invalid-amount": {
			given: args{
				owner:  owner,
				amount: amount * -1,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)
			},
			wantErr: ErrInvalidAmount,
		},
		"insufficient-balance": {
			given: args{
				owner:  owner,
				amount: amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    500,
				}, nil)
			},
			wantErr: ErrInsufficentBalance,
		},
		"multi-transaction-consumption-remaining": {
			given: args{
				owner:  owner,
				amount: -400.0,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "2000000",
						CreatedAt:     now.Add(10),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        100.0,
					},
					{
						TransactionId: "3000000",
						CreatedAt:     now.Add(50),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        300.0,
					},
				}

				debitTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: true,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     -400,
				}

				transaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     200,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    600,
				}, nil)
				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[1].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[2].TransactionId).Return(nil)
				deps.transRepoMock.On("Create", ctx, transaction).Return(nil)
				deps.transRepoMock.On("Create", ctx, debitTransaction).Return(nil)
			},
			wantErr: nil,
		},
		"multi-transaction-consumption-exact": {
			given: args{
				owner:  owner,
				amount: -400.0,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "2000000",
						CreatedAt:     now.Add(10),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "3000000",
						CreatedAt:     now.Add(50),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
				}

				debitTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: true,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     -400,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    600,
				}, nil)
				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[1].TransactionId).Return(nil)
				deps.transRepoMock.On("Create", ctx, debitTransaction).Return(nil)
			},
			wantErr: nil,
		},
		"multi-transaction-consumption-rollback": {
			given: args{
				owner:  owner,
				amount: -400.0,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "2000000",
						CreatedAt:     now.Add(10),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        100.0,
					},
					{
						TransactionId: "3000000",
						CreatedAt:     now.Add(50),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        300.0,
					},
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    600,
				}, nil)
				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[1].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[2].TransactionId).Return(repository.ErrTransactionNotFound)
				deps.transRepoMock.On("RollBackConsumed", ctx, []string{"1000000", "2000000"}).Return(nil)
			},
			wantErr: ErrFailedDebitOperation,
		},
	}

	for name, tcase := range scenarios {
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			service, deps := setupTransactionService(t)
			if tcase.doMocks != nil {
				tcase.doMocks(deps)
			}

			err := service.Withdraw(ctx, tcase.given.owner, tcase.given.amount)

			if tcase.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tcase.wantErr)
			}
		})
	}
}

func TestTransactionService_Pay(t *testing.T) {
	now := time.Now()
	setupClock(now)
	utils.SetSyncGoroutine()
	defer utils.ResetGoroutine()
	defer resetClock()

	var (
		ctx              = context.Background()
		owner            = "0001"
		receiver         = "0002"
		amount   float32 = 1000.0
	)

	type args struct {
		owner    string
		receiver string
		amount   float32
	}

	scenarios := map[string]struct {
		given   args
		doMocks func(deps transactionServiceDependencies)
		wantErr error
	}{
		"happy-path": {
			given: args{
				owner:    owner,
				receiver: receiver,
				amount:   amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        7000.0,
					},
				}

				debitTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: true,
					Owner:      owner,
					Sender:     owner,
					Receiver:   receiver,
					Amount:     -amount,
				}

				senderAdjTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      owner,
					Sender:     owner,
					Receiver:   owner,
					Amount:     7000 - amount,
				}

				receiverTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      receiver,
					Sender:     owner,
					Receiver:   receiver,
					Amount:     amount,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.accRepoMock.On("FindOne", ctx, receiver).Return(models.Account{
					AccountId: owner,
					Name:      "Jessica",
					LastName:  "Lourenco",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    7000,
				}, nil)

				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("Create", ctx, senderAdjTransaction).Return(nil)
				deps.transRepoMock.On("Create", ctx, debitTransaction).Return(nil)
				deps.transRepoMock.On("Create", ctx, receiverTransaction).Return(nil)
			},
			wantErr: nil,
		},
		"invalid-operation": {
			given: args{
				owner:    owner,
				receiver: owner,
				amount:   amount,
			},
			wantErr: ErrInvalidPaymentOp,
		},
		"invalid-owner": {
			given: args{
				owner:    owner,
				receiver: receiver,
				amount:   amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{}, repository.ErrAccountNotFound)
			},
			wantErr: repository.ErrAccountNotFound,
		},
		"invalid-receiver": {
			given: args{
				owner:    owner,
				receiver: receiver,
				amount:   amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)
				deps.accRepoMock.On("FindOne", ctx, receiver).Return(models.Account{}, repository.ErrAccountNotFound)
			},
			wantErr: repository.ErrAccountNotFound,
		},
		"insufficient-balance": {
			given: args{
				owner:    owner,
				receiver: receiver,
				amount:   amount,
			},
			doMocks: func(deps transactionServiceDependencies) {
				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.accRepoMock.On("FindOne", ctx, receiver).Return(models.Account{
					AccountId: owner,
					Name:      "Jessica",
					LastName:  "Lourenco",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    500,
				}, nil)
			},
			wantErr: ErrInsufficentBalance,
		},
		"multi-transaction-consumption": {
			given: args{
				owner:    owner,
				receiver: receiver,
				amount:   400.0,
			},
			doMocks: func(deps transactionServiceDependencies) {
				transactions := []models.Transaction{
					{
						TransactionId: "1000000",
						CreatedAt:     now,
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "2000000",
						CreatedAt:     now.Add(10),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
					{
						TransactionId: "3000000",
						CreatedAt:     now.Add(50),
						IsConsumed:    false,
						Owner:         owner,
						Sender:        owner,
						Receiver:      owner,
						Amount:        200.0,
					},
				}

				debitTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: true,
					Owner:      owner,
					Sender:     owner,
					Receiver:   receiver,
					Amount:     -400.0,
				}

				receiverTransaction := models.Transaction{
					CreatedAt:  now,
					IsConsumed: false,
					Owner:      receiver,
					Sender:     owner,
					Receiver:   receiver,
					Amount:     400.0,
				}

				deps.accRepoMock.On("FindOne", ctx, owner).Return(models.Account{
					AccountId: owner,
					Name:      "Shankar",
					LastName:  "Nakai",
				}, nil)

				deps.accRepoMock.On("FindOne", ctx, receiver).Return(models.Account{
					AccountId: owner,
					Name:      "Jessica",
					LastName:  "Lourenco",
				}, nil)

				deps.transRepoMock.On("GetBalance", ctx, owner).Return(models.Balance{
					AccountId: owner,
					Amount:    600,
				}, nil)

				deps.transRepoMock.On("FindAll", ctx, owner).Return(transactions, nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[0].TransactionId).Return(nil)
				deps.transRepoMock.On("MarkAsConsumed", ctx, transactions[1].TransactionId).Return(nil)
				deps.transRepoMock.On("Create", ctx, debitTransaction).Return(nil)
				deps.transRepoMock.On("Create", ctx, receiverTransaction).Return(nil)
			},
			wantErr: nil,
		},
	}

	for name, tcase := range scenarios {
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			service, deps := setupTransactionService(t)
			if tcase.doMocks != nil {
				tcase.doMocks(deps)
			}

			err := service.Pay(ctx, tcase.given.owner, tcase.given.receiver, tcase.given.amount)

			if tcase.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tcase.wantErr)
			}
		})
	}
}

type transactionServiceDependencies struct {
	transRepoMock *repository.MockTransactionRepo
	accRepoMock   *repository.MockAccountRepo
}

func setupTransactionService(t *testing.T) (*transactionServiceImpl, transactionServiceDependencies) {
	deps := transactionServiceDependencies{
		transRepoMock: repository.NewMockTransactionRepo(t),
		accRepoMock:   repository.NewMockAccountRepo(t),
	}

	return NewTransactionService(deps.transRepoMock, deps.accRepoMock), deps
}
