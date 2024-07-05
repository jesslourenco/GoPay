package models

type AmountReq struct {
	Amount float32 `json:"amount"`
}

type AccReq struct {
	Name     string `json:"name"`
	LastName string `json:"lastname"`
}
