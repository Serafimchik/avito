package models

type CoinTransaction struct {
	User   string `json:"user"`
	Amount int64  `json:"amount"`
}
