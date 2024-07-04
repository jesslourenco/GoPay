package main

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gopay/internal"
	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
	"github.com/gopay/internal/service"
)

func initDB() {
	// temp fake db for accounts and transactions
	models.Accounts["0001"] = &models.Account{
		AccountId: "0001",
		Name:      "Shankar",
		LastName:  "Nakai",
	}

	models.Accounts["0002"] = &models.Account{
		AccountId: "0002",
		Name:      "Jessica",
		LastName:  "Lourenco",
	}

	models.Accounts["0003"] = &models.Account{
		AccountId: "0003",
		Name:      "Caio",
		LastName:  "Henrique",
	}

	models.Accounts["0004"] = &models.Account{
		AccountId: "0004",
		Name:      "Karina",
		LastName:  "Domingues",
	}

	models.Transactions["1000000"] = &models.Transaction{
		TransactionId: "1000000",
		Owner:         "0001",
		Sender:        "0001",
		Receiver:      "0001",
		CreatedAt:     time.Now(),
		Amount:        7000.00,
		IsConsumed:    false,
	}

	models.Transactions["2000000"] = &models.Transaction{
		TransactionId: "2000000",
		Owner:         "0002",
		Sender:        "0002",
		Receiver:      "0002",
		CreatedAt:     time.Now(),
		Amount:        3000.00,
		IsConsumed:    false,
	}
}

func main() {
	transactionRepo := repository.NewTransactionRepo()
	accountRepo := repository.NewAccountRepo()
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo)

	apiHandler := internal.NewAPIHandler(transactionSvc)

	router := internal.Router(apiHandler)

	initDB()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msg("Server started at port :8080")

	log.
		Fatal().
		Err(http.ListenAndServe(":8080", router)).
		Msg("Server closed")
}
