package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/models"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

type AccountServiceInterface interface {
	CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error)
	CreateAccount(ctx context.Context) (*models.AccountResult, error)
	TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error)
	FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error)
	FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error)
	VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error)
	TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error)
	CheckAliasAddress(ctx context.Context, alias string) (*dataserviceapi.AliasAddress, error)
}

type AccountService struct {
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
func (as *AccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	var r models.TrackStatusResult

	ep, err := url.JoinPath(config.TrackURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *AccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	var balanceResult models.BalanceResult

	ep, err := url.JoinPath(config.BalanceURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &balanceResult)
	return &balanceResult, err
}

// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	var r models.AccountResult
	// Create a new request
	req, err := http.NewRequest("POST", config.CreateAccountURL, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FetchVouchers retrieves the token holdings for a given public key from the data indexer API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *AccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	var r struct {
		Holdings []dataserviceapi.TokenHoldings `json:"holdings"`
	}

	ep, err := url.JoinPath(config.VoucherHoldingsURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return r.Holdings, nil
}

// FetchTransactions retrieves the last 10 transactions for a given public key from the data indexer API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *AccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	var r struct {
		Transfers []dataserviceapi.Last10TxResponse `json:"transfers"`
	}

	ep, err := url.JoinPath(config.VoucherTransfersURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return r.Transfers, nil
}

// VoucherData retrieves voucher metadata from the data indexer API endpoint.
// Parameters:
//   - address: The voucher address.
func (as *AccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	var r struct {
		TokenDetails models.VoucherDataResult `json:"tokenDetails"`
	}

	ep, err := url.JoinPath(config.VoucherDataURL, address)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	return &r.TokenDetails, err
}

// TokenTransfer creates a new token transfer in the custodial system.
// Returns:
//   - *models.TokenTransferResponse: A pointer to an TokenTransferResponse struct containing the trackingId.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	var r models.TokenTransferResponse

	// Create request payload
	payload := map[string]string{
		"amount":       amount,
		"from":         from,
		"to":           to,
		"tokenAddress": tokenAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create a new request
	req, err := http.NewRequest("POST", config.TokenTransferURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// CheckAliasAddress retrieves the address of an alias from the API endpoint.
// Parameters:
//   - alias: The alias of the user.
func (as *AccountService) CheckAliasAddress(ctx context.Context, alias string) (*dataserviceapi.AliasAddress, error) {
	var r dataserviceapi.AliasAddress

	ep, err := url.JoinPath(config.CheckAliasURL, alias)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	return &r, err
}

func doRequest(ctx context.Context, req *http.Request, rcpt any) (*api.OKResponse, error) {
	var okResponse api.OKResponse
	var errResponse api.ErrResponse

	req.Header.Set("Authorization", "Bearer "+config.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	logRequestDetails(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to make %s request to endpoint: %s with reason: %s", req.Method, req.URL, err.Error())
		errResponse.Description = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Received response for %s: Status Code: %d | Content-Type: %s", req.URL, resp.StatusCode, resp.Header.Get("Content-Type"))
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

	v, err := json.Marshal(okResponse.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(v, &rcpt)
	return &okResponse, err
}

func logRequestDetails(req *http.Request) {
	var bodyBytes []byte
	contentType := req.Header.Get("Content-Type")
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Error reading request body: %s", err)
			return
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		bodyBytes = []byte("-")
	}

	log.Printf("URL: %s | Content-Type: %s | Method: %s| Request Body: %s", req.URL, contentType, req.Method, string(bodyBytes))
}
