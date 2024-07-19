package repository

import (
	"context"
	"database/sql"

	"github.com/gopay/internal/models"
)

const (
	createTransQ = `
	INSERT INTO transactions 
	(owner, sender, receiver, created_at, amount, is_consumed) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	`

	findAllTransQ = `
	SELECT transaction_id, owner, sender, receiver, created_at, amount, is_consumed
	FROM transactions
	ORDER BY created_at ASC
	`

	findOneTransQ = `
	SELECT transaction_id, owner, sender, receiver, created_at, amount, is_consumed
	FROM transactions
	WHERE transaction_id = $1
	`

	getBalanceQ = `
	SELECT SUM(amount)
	FROM transactions
	WHERE is_consumed = false
	AND owner = $1
	`

	updateAsConsumedQ = `
	UPDATE transactions
	SET is_consumed = true
	WHERE transaction_id = $1
	`
)

type TransactionRepoPsql interface {
	FindAll(ctx context.Context, accId string) ([]models.Transaction, error)
	FindOne(ctx context.Context, id string) (models.Transaction, error)
	Create(ctx context.Context, transaction models.Transaction) error
	MarkAsConsumed(ctx context.Context, id string) error
	GetBalance(ctx context.Context, id string) (models.Balance, error)
	RollBackConsumed(ctx context.Context, tConsumed []string) error
}

var _ TransactionRepoPsql = (*transactionRepoPsqlImpl)(nil)

type transactionRepoPsqlImpl struct {
	psql *sql.DB
}

func NewTransactionRepoPsql(db *sql.DB) *transactionRepoPsqlImpl {
	return &transactionRepoPsqlImpl{
		psql: db,
	}
}

func (r *transactionRepoPsqlImpl) FindAll(ctx context.Context, accId string) ([]models.Transaction, error) {
	transactions := []models.Transaction{}

	rows, err := r.psql.Query(findAllTransQ)
	if err != nil {
		return transactions, err
	}
	defer rows.Close()

	for rows.Next() {
		t := models.Transaction{}
		rows.Scan(&t.TransactionId, &t.Owner, &t.Sender, &t.Receiver, &t.CreatedAt, &t.Amount, &t.IsConsumed)
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *transactionRepoPsqlImpl) FindOne(ctx context.Context, id string) (models.Transaction, error) {
	t := models.Transaction{}

	row := r.psql.QueryRow(findOneTransQ, id)
	err := row.Scan(&t.TransactionId, &t.Owner, &t.Sender, &t.Receiver, &t.CreatedAt, &t.Amount, &t.IsConsumed)
	if err == sql.ErrNoRows {
		return t, ErrTransactionNotFound
	}
	if err != nil {
		return models.Transaction{}, err
	}

	return t, nil
}

func (r *transactionRepoPsqlImpl) Create(ctx context.Context, transaction models.Transaction) error {
	if transaction.Sender == "" {
		return ErrMissingSenderField
	}
	if transaction.Receiver == "" {
		return ErrMissingReceiverField
	}

	if transaction.Owner == "" {
		return ErrMissingOwnerField
	}

	if transaction.Amount == 0 {
		return ErrZeroAmount
	}

	_, err := r.psql.Exec(createTransQ, transaction.Owner, transaction.Sender, transaction.Receiver, transaction.CreatedAt, transaction.Amount, transaction.IsConsumed)
	if err != nil {
		return err
	}

	return nil
}

func (r *transactionRepoPsqlImpl) MarkAsConsumed(ctx context.Context, id string) error {
	res, err := r.psql.Exec(updateAsConsumedQ, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

func (r *transactionRepoPsqlImpl) GetBalance(ctx context.Context, id string) (models.Balance, error) {
	balance := models.Balance{
		AccountId: id,
		Amount:    0.0,
	}

	row := r.psql.QueryRow(getBalanceQ, id)
	err := row.Scan(&balance.Amount)
	if err != nil {
		return balance, err
	}

	if balance.Amount < 0 {
		return balance, ErrNegativeBalance
	}

	return balance, nil
}

func (r *transactionRepoPsqlImpl) RollBackConsumed(ctx context.Context, tConsumed []string) error {
	return nil
}
