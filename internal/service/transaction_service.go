package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
	"github.com/gopay/internal/utils"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidAmount        = errors.New("amount cannot be less or equal to zero")
	ErrInsufficentBalance   = errors.New("insufficient balance")
	ErrFailedDebitOperation = errors.New("debit operation  unsuccessful ")
	ErrFaileCreditOperation = errors.New("credit operation  unsuccessful ")
	ErrInvalidPaymentOp     = errors.New("sender and receiver accounts must be different")
)

var nowOriginal = func() time.Time {
	return time.Now()
}
var clockNow = nowOriginal

func setupClock(value time.Time) {
	clockNow = func() time.Time {
		return value
	}
}

func resetClock() {
	clockNow = nowOriginal
}

type TransactionService interface {
	Deposit(ctx context.Context, owner string, amount float32) error
	Withdraw(ctx context.Context, owner string, amount float32) error
	Pay(ctx context.Context, owner string, receiver string, amount float32) error
	GetTransaction(ctx context.Context, id string) (models.Transaction, error)
	GetAllTransactions(ctx context.Context, accId string) ([]models.Transaction, error)
	GetBalance(ctx context.Context, accId string) (models.Balance, error)
}

var _ TransactionService = (*transactionServiceImpl)(nil)

type transactionServiceImpl struct {
	transactionRepo repository.TransactionRepo
	accountRepo     repository.AccountRepo
}

func NewTransactionService(transactionRepo repository.TransactionRepo, accountRepo repository.AccountRepo) *transactionServiceImpl {
	return &transactionServiceImpl{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

func (r *transactionServiceImpl) GetBalance(ctx context.Context, accId string) (models.Balance, error) {
	_, err := r.accountRepo.FindOne(ctx, accId)
	if err != nil {
		return models.Balance{}, err
	}

	return r.transactionRepo.GetBalance(ctx, accId)
}

func (r *transactionServiceImpl) GetAllTransactions(ctx context.Context, accId string) ([]models.Transaction, error) {
	_, err := r.accountRepo.FindOne(ctx, accId)
	if err != nil {
		return []models.Transaction{}, err
	}

	return r.transactionRepo.FindAll(ctx, accId)
}

func (r *transactionServiceImpl) GetTransaction(ctx context.Context, id string) (models.Transaction, error) {
	return r.transactionRepo.FindOne(ctx, id)
}

func (r *transactionServiceImpl) Deposit(ctx context.Context, owner string, amount float32) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	_, err := r.accountRepo.FindOne(ctx, owner)
	if err != nil {
		return err
	}

	return r.credit(ctx, owner, owner, owner, amount)
}

func (r *transactionServiceImpl) Withdraw(ctx context.Context, owner string, amount float32) error {
	_, err := r.accountRepo.FindOne(ctx, owner)
	if err != nil {
		return err
	}

	consumed, err := r.debit(ctx, owner, owner, amount)
	if err != nil {
		return err
	}

	transaction := models.Transaction{
		CreatedAt:  clockNow(),
		IsConsumed: true,
		Owner:      owner,
		Sender:     owner,
		Receiver:   owner,
		Amount:     amount,
	}

	err = r.transactionRepo.Create(ctx, transaction)
	if err != nil {
		go func() {
			err := utils.Retry(func() error {
				return r.transactionRepo.RollBackConsumed(ctx, consumed)
			}, "rollback of MarkAsConsumed")
			log.Error().Err(err)
		}()
		log.Error().Err(err)
		return ErrFailedDebitOperation
	}

	return nil
}

func (r *transactionServiceImpl) Pay(ctx context.Context, owner string, receiver string, amount float32) error {
	if owner == receiver {
		return ErrInvalidPaymentOp
	}

	_, err := r.accountRepo.FindOne(ctx, owner)
	if err != nil {
		return err
	}

	_, err = r.accountRepo.FindOne(ctx, receiver)
	if err != nil {
		return err
	}

	amount = float32(math.Abs(float64(amount)))

	consumed, err := r.debit(ctx, owner, owner, -amount)
	if err != nil {
		return err
	}

	transaction := models.Transaction{
		CreatedAt:  clockNow(),
		IsConsumed: true,
		Owner:      owner,
		Sender:     owner,
		Receiver:   receiver,
		Amount:     -amount,
	}

	err = r.transactionRepo.Create(ctx, transaction)
	if err != nil {
		go func() {
			err := utils.Retry(func() error {
				return r.transactionRepo.RollBackConsumed(ctx, consumed)
			}, "rollback of MarkAsConsumed")
			log.Error().Err(err)
		}()
		log.Error().Err(err)
		return ErrFailedDebitOperation
	}

	err = r.credit(ctx, receiver, owner, receiver, amount)
	if err != nil {
		go func() {
			err := utils.Retry(func() error {
				return r.transactionRepo.RollBackConsumed(ctx, consumed)
			}, "rollback of MarkAsConsumed")
			log.Error().Err(err)
		}()
		log.Error().Err(err)
		return ErrFaileCreditOperation
	}

	return nil
}

func (r *transactionServiceImpl) credit(ctx context.Context, owner string, sender string, receiver string, amount float32) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	transaction := models.Transaction{
		CreatedAt:  clockNow(),
		IsConsumed: false,
		Owner:      owner,
		Sender:     sender,
		Receiver:   receiver,
		Amount:     amount,
	}

	return r.transactionRepo.Create(ctx, transaction)
}

func (r *transactionServiceImpl) debit(ctx context.Context, owner string, receiver string, amount float32) ([]string, error) {
	if amount >= 0 {
		return []string{}, ErrInvalidAmount
	}

	balance, err := r.transactionRepo.GetBalance(ctx, owner)
	if err != nil {
		return []string{}, err
	}

	if (balance.Amount + float64(amount)) < 0 {
		return []string{}, ErrInsufficentBalance
	}

	transactions, err := r.transactionRepo.FindAll(ctx, owner)
	if err != nil {
		return []string{}, err
	}

	debit := (-1) * amount
	transConsumed := []string{}
	for _, t := range transactions {
		err = r.transactionRepo.MarkAsConsumed(ctx, t.TransactionId)
		if err != nil {
			utils.Go(func() {
				err := utils.Retry(func() error {
					return r.transactionRepo.RollBackConsumed(ctx, transConsumed)
				}, "rollback of MarkAsConsumed")
				log.Error().Err(err)
			})
			log.Error().Err(err)
			return []string{}, ErrFailedDebitOperation
		}

		transConsumed = append(transConsumed, t.TransactionId)
		remaining := t.Amount - debit

		if remaining == 0 {
			break
		}

		if remaining < 0 {
			debit = debit - t.Amount
			continue
		}

		if remaining > 0 {
			err := r.credit(ctx, owner, owner, receiver, t.Amount-debit)
			if err != nil {
				go func() {
					err := utils.Retry(func() error {
						return r.transactionRepo.RollBackConsumed(ctx, transConsumed)
					}, "rollback of MarkAsConsumed")
					log.Error().Err(err)
				}()
				log.Error().Err(err)
				return []string{}, ErrFailedDebitOperation
			}

			break
		}
	}

	return transConsumed, nil
}
