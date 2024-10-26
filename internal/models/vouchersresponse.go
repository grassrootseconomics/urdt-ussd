package models

// VoucherHoldingResponse represents a single voucher holding
type VoucherHoldingResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		Holdings []struct {
			ContractAddress string `json:"contractAddress"`
			TokenSymbol     string `json:"tokenSymbol"`
			TokenDecimals   string `json:"tokenDecimals"`
			Balance         string `json:"balance"`
		} `json:"holdings"`
	} `json:"result"`
}
