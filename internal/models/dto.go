package models

type AmountReq struct {
	Amount float32 `json:"amount"`
}

type AccReq struct {
	Name     string `json:"name"`
	LastName string `json:"lastname"`
}

type PayReq struct {
	Receiver string  `json:"receiver"`
	Amount   float32 `json:"amount"`
}
