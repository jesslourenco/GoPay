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
	if err != nil {
		log.Fatal().Msg("ping failed")
	}

	log.Info().Msg("Pong")
	log.Info().Msg("PostgreSql connected successfully...")

	transactionRepo := repository.NewTransactionRepoPsql(db)
	accountRepo := repository.NewAccountRepoPsql(db)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo)
	accountSvc := service.NewAccountService(accountRepo)

	apiHandler := internal.NewAPIHandler(transactionSvc, accountSvc)

	router := internal.Router(apiHandler)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msgf("Server started at port %s", config.ServerAddress)

	log.
		Fatal().
		Err(http.ListenAndServe(config.ServerAddress, router)).
		Msg("Server closed")
}
