package testservice

import (
	"context"
	"encoding/json"

	"git.grassecon.net/urdt/ussd/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

type TestAccountService struct {
}

func (tas *TestAccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	return &models.AccountResult{
		TrackingId: "075ccc86-f6ef-4d33-97d5-e91cfb37aa0d",
		PublicKey:  "0x623EFAFa8868df4B934dd12a8B26CB3Dd75A7AdD",
	}, nil
}

func (tas *TestAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	balanceResponse := &models.BalanceResult{
		Balance: "0.003 CELO",
		Nonce:   json.Number("0"),
	}
	return balanceResponse, nil
}

func (tas *TestAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	return &models.TrackStatusResult{
		Active: true,
	}, nil
}

func (tas *TestAccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	return []dataserviceapi.TokenHoldings{
		dataserviceapi.TokenHoldings{
			ContractAddress: "0x6CC75A06ac72eB4Db2eE22F781F5D100d8ec03ee",
			TokenSymbol:     "SRF",
			TokenDecimals:   "6",
			Balance:         "2745987",
		},
	}, nil
}

func (tas *TestAccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	return []dataserviceapi.Last10TxResponse{}, nil
}

func (m TestAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	return &models.VoucherDataResult{}, nil
}

func (tas *TestAccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	return &models.TokenTransferResponse{
		TrackingId: "e034d147-747d-42ea-928d-b5a7cb3426af",
	}, nil
}

func (m TestAccountService) CheckAliasAddress(ctx context.Context, alias string) (*dataserviceapi.AliasAddress, error) {
	return &dataserviceapi.AliasAddress{}, nil
}
