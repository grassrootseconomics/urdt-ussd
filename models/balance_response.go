package models

import "encoding/json"

type BalanceResult struct {
	Balance string      `json:"balance"`
	Nonce   json.Number `json:"nonce"`
}
