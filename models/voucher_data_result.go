package models

type VoucherDataResult struct {
	TokenName      string `json:"tokenName"`
	TokenSymbol    string `json:"tokenSymbol"`
	TokenDecimals  int    `json:"tokenDecimals"`
	SinkAddress    string `json:"sinkAddress"`
	TokenCommodity string `json:"tokenCommodity"`
	TokenLocation  string `json:"tokenLocation"`
}
