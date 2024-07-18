package repository

import (
	"context"
	"database/sql"

	"github.com/gopay/internal/models"
)

const (
	createAccQ = `
	INSERT INTO accounts 
	(name, last_name) 
	VALUES ($1, $2) 
	RETURNING account_id
	`

	findAllAccsQ = `
	SELECT account_id, name, last_name 
	FROM accounts
	`

	findOneAccQ = `
	SELECT account_id, name, last_name
	FROM accounts
	WHERE account_id = $1
	`
)

type AccountRepoPsql interface {
	FindAll(ctx context.Context) ([]models.Account, error)
	FindOne(ctx context.Context, id string) (models.Account, error)
	Create(ctx context.Context, name string, lastname string) (string, error)
}

var _ AccountRepoPsql = (*accountRepoPsqlImpl)(nil)

type accountRepoPsqlImpl struct {
	psql *sql.DB
}

func NewAccountRepoPsql(db *sql.DB) *accountRepoPsqlImpl {
	return &accountRepoPsqlImpl{
		psql: db,
	}
}

func (r *accountRepoPsqlImpl) FindAll(_ context.Context) ([]models.Account, error) {
	accs := []models.Account{}

	rows, err := r.psql.Query(findAllAccsQ)
	if err != nil {
		return accs, err
	}
	defer rows.Close()

	for rows.Next() {
		acc := models.Account{}
		rows.Scan(&acc.AccountId, &acc.Name, &acc.LastName)
		accs = append(accs, acc)
	}

	return accs, nil
}

func (r *accountRepoPsqlImpl) FindOne(ctx context.Context, id string) (models.Account, error) {
	acc := models.Account{}

	row := r.psql.QueryRow(findOneAccQ, id)
	err := row.Scan(&acc.AccountId, &acc.Name, &acc.LastName)
	if err == sql.ErrNoRows {
		return acc, ErrAccountNotFound
	}
	if err != nil {
		return models.Account{}, err
	}

	return acc, nil
}

func (r *accountRepoPsqlImpl) Create(ctx context.Context, name string, lastname string) (string, error) {
	if name == "" || lastname == "" {
		return "", ErrMissingParams
	}

	var id string

	row := r.psql.QueryRow(createAccQ, name, lastname)
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}
