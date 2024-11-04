package models

import dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"

//type VoucherHoldingResponse struct {
//	Ok          bool          `json:"ok"`
//	Description string        `json:"description"`
//	Result      VoucherResult `json:"result"`
//}

// VoucherResult holds the list of token holdings
type VoucherResult struct {
	Holdings []dataserviceapi.TokenHoldings `json:"holdings"`
}

type VoucherDataResult struct {
	TokenName	string `json:"tokenName"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimals   string `json:"tokenDecimals"`
	SinkAddress	string `json:"sinkAddress"`
}
