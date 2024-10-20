package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

type AccountServiceInterface interface {
	CheckBalance(publicKey string) (*models.BalanceResponse, error)
	CreateAccount() (*models.AccountResponse, error)
	CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error)
}

type AccountService struct {
}

type TestAccountService struct {
}

// CheckAccountStatus retrieves the status of an account transaction based on the provided tracking ID.
//
// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	resp, err := http.Get(config.TrackStatusURL + trackingId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var trackResp models.TrackStatusResponse
	err = json.Unmarshal(body, &trackResp)
	if err != nil {
		return nil, err
	}
	return &trackResp, nil
}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *AccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {
	resp, err := http.Get(config.BalanceURL + publicKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var balanceResp models.BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return nil, err
	}
	return &balanceResp, nil
}

// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) CreateAccount() (*models.AccountResponse, error) {
	resp, err := http.Post(config.CreateAccountURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var accountResp models.AccountResponse
	err = json.Unmarshal(body, &accountResp)
	if err != nil {
		return nil, err
	}
	return &accountResp, nil
}

func (tas *TestAccountService) CreateAccount() (*models.AccountResponse, error) {
	return &models.AccountResponse{
		Ok: true,
		Result: struct {
			CustodialId json.Number `json:"custodialId"`
			PublicKey   string      `json:"publicKey"`
			TrackingId  string      `json:"trackingId"`
		}{
			CustodialId: json.Number("182"),
			PublicKey:   "0x48ADca309b5085852207FAaf2816eD72B52F527C",
			TrackingId:  "28ebe84d-b925-472c-87ae-bbdfa1fb97be",
		},
	}, nil
}

func (tas *TestAccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {

	balanceResponse := &models.BalanceResponse{
		Ok: true,
		Result: struct {
			Balance string      `json:"balance"`
			Nonce   json.Number `json:"nonce"`
		}{
			Balance: "0.003 CELO",
			Nonce:   json.Number("0"),
		},
	}

	return balanceResponse, nil
}

func (tas *TestAccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	trackResponse := &models.TrackStatusResponse{
		Ok: true,
		Result: struct {
			Transaction struct {
				CreatedAt     time.Time   "json:\"createdAt\""
				Status        string      "json:\"status\""
				TransferValue json.Number "json:\"transferValue\""
				TxHash        string      "json:\"txHash\""
				TxType        string      "json:\"txType\""
			}
		}{
			Transaction: models.Transaction{
				CreatedAt:     time.Now(),
				Status:        "SUCCESS",
				TransferValue: json.Number("0.5"),
				TxHash:        "0x123abc456def",
				TxType:        "transfer",
			},
		},
	}
	return trackResponse, nil
}
