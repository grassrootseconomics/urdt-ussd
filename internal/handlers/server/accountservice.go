package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
)

var (
	okResponse  api.OKResponse
	errResponse api.ErrResponse
)

type AccountServiceInterface interface {
	CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error)
	CreateAccount(ctx context.Context) (*api.OKResponse, error)
	CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error)
	TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error)
	FetchVouchers(ctx context.Context, publicKey string) (*models.VoucherHoldingResponse, error)
}

type AccountService struct {
}

type TestAccountService struct {
}

// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil
func (as *AccountService) CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error) {
	resp, err := http.Get(config.BalanceURL + trackingId)
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

func (as *AccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error) {
	var err error
	// Construct the URL with the path parameter
	url := fmt.Sprintf("%s/%s", config.TrackURL, publicKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GE-KEY", "xd")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.Description)
	}
	err = json.Unmarshal([]byte(body), &okResponse)
	if err != nil {
		return nil, err
	}
	if len(okResponse.Result) == 0 {
		return nil, errors.New("Empty api result")
	}
	return &okResponse, nil

}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *AccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error) {
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
func (as *AccountService) CreateAccount(ctx context.Context) (*api.OKResponse, error) {
	var err error

	// Create a new request
	req, err := http.NewRequest("POST", config.CreateAccountURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GE-KEY", "xd")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.Description)
	}
	err = json.Unmarshal([]byte(body), &okResponse)
	if err != nil {
		return nil, err
	}
	if len(okResponse.Result) == 0 {
		return nil, errors.New("Empty api result")
	}
	return &okResponse, nil
}

// FetchVouchers retrieves the token holdings for a given public key from the custodial holdings API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *AccountService) FetchVouchers(ctx context.Context, publicKey string) (*models.VoucherHoldingResponse, error) {
	file, err := os.Open("sample_tokens.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var holdings models.VoucherHoldingResponse

	if err := json.NewDecoder(file).Decode(&holdings); err != nil {
		return nil, err
	}
	return &holdings, nil
}

func (tas *TestAccountService) CreateAccount(ctx context.Context) (*api.OKResponse, error) {
	return &api.OKResponse{
		Ok:          true,
		Description: "Account creation request received successfully",
		Result:      map[string]any{"publicKey": "0x48ADca309b5085852207FAaf2816eD72B52F527C", "trackingId": "28ebe84d-b925-472c-87ae-bbdfa1fb97be"},
	}, nil

}

func (tas *TestAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error) {
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

func (tas *TestAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error) {
	return &api.OKResponse{
		Ok:          true,
		Description: "Account creation succeeded",
		Result: map[string]any{
			"active": true,
		},
	}, nil
}

func (tas *TestAccountService) CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error) {
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

func (tas *TestAccountService) FetchVouchers(ctx context.Context, publicKey string) (*models.VoucherHoldingResponse, error) {
	return &models.VoucherHoldingResponse{
		Ok: true,
		Result: struct {
			Holdings []struct {
				ContractAddress string `json:"contractAddress"`
				TokenSymbol     string `json:"tokenSymbol"`
				TokenDecimals   string `json:"tokenDecimals"`
				Balance         string `json:"balance"`
			} `json:"holdings"`
		}{
			Holdings: []struct {
				ContractAddress string `json:"contractAddress"`
				TokenSymbol     string `json:"tokenSymbol"`
				TokenDecimals   string `json:"tokenDecimals"`
				Balance         string `json:"balance"`
			}{
				{
					ContractAddress: "0x6CC75A06ac72eB4Db2eE22F781F5D100d8ec03ee",
					TokenSymbol:     "SRF",
					TokenDecimals:   "6",
					Balance:         "2745987",
				},
			},
		},
	}, nil
}
