package main

import (
	"database/sql"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gopay/internal"
	"github.com/gopay/internal/repository"
	"github.com/gopay/internal/service"
	"github.com/gopay/internal/utils"
	_ "github.com/lib/pq"
)

/*func initDB() {
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
}*/

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("could not loadconfig: %v", err)
	}

	db, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal().Msgf("Could not connect to database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		log.Info().Msg("Pong")
		log.Info().Msg("PostgreSql connected successfully...")
	} else {
		log.Fatal().Msg("ping failed")
	}

	transactionRepo := repository.NewTransactionRepo()
	// accountRepo := repository.NewAccountRepo()
	accountRepo := repository.NewAccountRepoPsql(db)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo)
	accountSvc := service.NewAccountService(accountRepo)

	apiHandler := internal.NewAPIHandler(transactionSvc, accountSvc)

	router := internal.Router(apiHandler)

	// initDB()
	// initAccounts(accountRepo)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msgf("Server started at port %s", config.ServerAddress)

	log.
		Fatal().
		Err(http.ListenAndServe(config.ServerAddress, router)).
		Msg("Server closed")
}

/*func initAccounts(accountRepo repository.AccountRepo) {
	accounts := []models.Account{
		{
			Name:     "Shankar",
			LastName: "Nakai",
		},
		{
			Name:     "Jessica",
			LastName: "Lourenco",
		},
		{
			Name:     "Caio",
			LastName: "Henrique",
		},
		{
			Name:     "Karina",
			LastName: "Domingues",
		},
	}

	for _, acc := range accounts {
		id, err := accountRepo.Create(context.Background(), acc.Name, acc.LastName)
		if err != nil {
			log.Err(err).Msg("something went wrong")
		}
		log.Info().Msg(fmt.Sprintf("%s => %s", id, acc.Name))
	}
}*/
