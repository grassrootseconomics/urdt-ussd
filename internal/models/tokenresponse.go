package models

type ApiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      Result `json:"result"`
}

type Result struct {
	Holdings []Holding `json:"holdings"`
}

type Holding struct {
	ContractAddress string `json:"contractAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimals   string `json:"tokenDecimals"`
	Balance         string `json:"balance"`
}
