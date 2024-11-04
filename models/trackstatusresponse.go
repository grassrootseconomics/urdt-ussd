package models

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	CreatedAt     time.Time   `json:"createdAt"`
	Status        string      `json:"status"`
	TransferValue json.Number `json:"transferValue"`
	TxHash        string      `json:"txHash"`
	TxType        string      `json:"txType"`
}

type TrackStatusResult struct {
	Active	bool `json:"active"`
}
