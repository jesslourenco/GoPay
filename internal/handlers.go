package internal

import (
	"errors"
	"io"
	"net/http"

	"github.com/gopay/internal/models"
	"github.com/gopay/internal/repository"
	"github.com/gopay/internal/service"
	"github.com/gopay/internal/utils"
	jsoniter "github.com/json-iterator/go"

	"github.com/rs/zerolog/log"

	"github.com/julienschmidt/httprouter"
)

const (
	AccountIdParam     = "account-id"
	TransactionIdParam = "transaction-id"
	OneMegabyte        = 1048576
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrAccountNotFound     = errors.New("account not found")
	ErrReceiverNotFound    = errors.New("receiver account not found")
	ErrSenderNotFound      = errors.New("sender account not found")
	ErrInsufficentBalance  = errors.New("insufficient balance")
)

type HandlerRegister interface {
	Register(router *httprouter.Router)
}

type apiHandler struct {
	transactionSvc service.TransactionService
	accountSvc     service.AccountService
}

func NewAPIHandler(transactionSvc service.TransactionService, accountSvc service.AccountService) *apiHandler {
	return &apiHandler{
		transactionSvc: transactionSvc,
		accountSvc:     accountSvc,
	}
}

func (h *apiHandler) Register(router *httprouter.Router) {
	router.Handle(http.MethodGet, "/", h.Index)
	router.Handle(http.MethodGet, "/accounts", h.GetAllAccounts)
	router.Handle(http.MethodGet, "/accounts/:account-id", h.GetAccount)
	router.Handle(http.MethodPost, "/accounts", h.PostAccount)
	router.Handle(http.MethodGet, "/accounts/:account-id/transactions", h.GetAllTransactions)
	router.Handle(http.MethodGet, "/transactions/:transaction-id", h.GetTransaction)
	router.Handle(http.MethodPost, "/accounts/:account-id/deposit", h.Deposit)
	router.Handle(http.MethodPost, "/accounts/:account-id/withdraw", h.Withdraw)
	router.Handle(http.MethodGet, "/accounts/:account-id/balance", h.GetBalance)
}

func (h *apiHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=UTF8")

	res, err := jsoniter.Marshal("Welcome to GoPay!")
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}

func (h *apiHandler) GetAllAccounts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	accounts, err := h.accountSvc.GetAllAccounts(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAllAccounts")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := jsoniter.Marshal(&accounts)
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAllAccounts")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}

func (h *apiHandler) GetAccount(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName(AccountIdParam)

	account, err := h.accountSvc.GetAccount(r.Context(), id)
	if err == ErrAccountNotFound {
		log.Error().Err(err).Msg("Handler::GetAccount")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAccount")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := jsoniter.Marshal(&account)
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAccount")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}

func (h *apiHandler) PostAccount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := io.ReadAll(io.LimitReader(r.Body, OneMegabyte))
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer r.Body.Close()

	account := models.AccReq{}
	err = jsoniter.Unmarshal(body, &account)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	_, err = h.accountSvc.CreateAccount(r.Context(), account.Name, account.LastName)
	if err == repository.ErrMissingParams {
		log.Error().Err(err).Msg("Handler::PostAccount")
		utils.ErrorWithMessage(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::PostAccount")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusCreated, nil)
}

func (h *apiHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	accountId := params.ByName(AccountIdParam)

	transactions, err := h.transactionSvc.GetAllTransactions(r.Context(), accountId)
	if err == repository.ErrAccountNotFound {
		log.Error().Err(err).Msg("Handler::GetAllTransactions")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAllTransactions")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := jsoniter.Marshal(&transactions)
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetAllTransactions")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}

func (h *apiHandler) GetTransaction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName(TransactionIdParam)

	transaction, err := h.transactionSvc.GetTransaction(r.Context(), id)

	if err == repository.ErrTransactionNotFound {
		log.Error().Err(err).Msg("Handler::GetTransaction")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::GetTransaction")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := jsoniter.Marshal(&transaction)
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetTransaction")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}

func (h *apiHandler) Deposit(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	owner := params.ByName(AccountIdParam)

	body, err := io.ReadAll(io.LimitReader(r.Body, OneMegabyte))
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer r.Body.Close()

	amount := models.AmountReq{}
	err = jsoniter.Unmarshal(body, &amount)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	err = h.transactionSvc.Deposit(r.Context(), owner, amount.Amount)
	if err == service.ErrInvalidAmount {
		log.Error().Err(err).Msg("Handler::Deposit")
		utils.ErrorWithMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	if err == repository.ErrAccountNotFound {
		log.Error().Err(err).Msg("Handler::Deposit")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::Deposit")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusCreated, nil)
}

func (h *apiHandler) Withdraw(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	owner := params.ByName(AccountIdParam)

	body, err := io.ReadAll(io.LimitReader(r.Body, OneMegabyte))
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer r.Body.Close()

	amount := models.AmountReq{}
	err = jsoniter.Unmarshal(body, &amount)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		utils.ErrorWithMessage(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	err = h.transactionSvc.Withdraw(r.Context(), owner, amount.Amount)
	if err == service.ErrInvalidAmount {
		log.Error().Err(err).Msg("Handler::Withdraw")
		utils.ErrorWithMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	if err == repository.ErrAccountNotFound {
		log.Error().Err(err).Msg("Handler::Withdraw")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err == service.ErrInsufficentBalance {
		log.Error().Err(err).Msg("Handler::Withdraw")
		utils.ErrorWithMessage(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::Withdraw")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusCreated, nil)
}

func (h *apiHandler) GetBalance(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	accountId := params.ByName(AccountIdParam)

	balance, err := h.transactionSvc.GetBalance(r.Context(), accountId)
	if err == repository.ErrAccountNotFound {
		log.Error().Err(err).Msg("Handler::GetBalance")
		utils.ErrorWithMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Handler::GetBalance")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := jsoniter.Marshal(&balance)
	if err != nil {
		log.Error().Err(err).Msg("Handler::GetBalance")
		utils.ErrorWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WithPayload(w, http.StatusOK, res)
}
