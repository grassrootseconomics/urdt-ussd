package config

import (
	"net/url"

	"git.grassecon.net/urdt/ussd/initializers"
)

const (
	createAccountPath          = "/api/v2/account/create"
	trackStatusPath            = "/api/track"
	balancePathPrefix          = "/api/account"
	trackPath                  = "/api/v2/account/status"
	tokenTransferPrefix        = "/api/v2/token/transfer"
	voucherHoldingsPathPrefix  = "/api/v1/holdings"
	voucherTransfersPathPrefix = "/api/v1/transfers/last10"
	voucherDataPathPrefix      = "/api/v1/token"
	AliasPrefix                = "api/v1/alias"
)

var (
	custodialURLBase string
	dataURLBase      string
	BearerToken      string
)

var (
	CreateAccountURL    string
	TrackStatusURL      string
	BalanceURL          string
	TrackURL            string
	TokenTransferURL    string
	VoucherHoldingsURL  string
	VoucherTransfersURL string
	VoucherDataURL      string
	CheckAliasURL       string
)

func setBase() error {
	var err error

	custodialURLBase = initializers.GetEnv("CUSTODIAL_URL_BASE", "http://localhost:5003")
	dataURLBase = initializers.GetEnv("DATA_URL_BASE", "http://localhost:5006")
	BearerToken = initializers.GetEnv("BEARER_TOKEN", "")

	_, err = url.JoinPath(custodialURLBase, "/foo")
	if err != nil {
		return err
	}
	_, err = url.JoinPath(dataURLBase, "/bar")
	if err != nil {
		return err
	}
	return nil
}

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() error {
	err := setBase()
	if err != nil {
		return err
	}
	CreateAccountURL, _ = url.JoinPath(custodialURLBase, createAccountPath)
	TrackStatusURL, _ = url.JoinPath(custodialURLBase, trackStatusPath)
	BalanceURL, _ = url.JoinPath(custodialURLBase, balancePathPrefix)
	TrackURL, _ = url.JoinPath(custodialURLBase, trackPath)
	TokenTransferURL, _ = url.JoinPath(custodialURLBase, tokenTransferPrefix)
	VoucherHoldingsURL, _ = url.JoinPath(dataURLBase, voucherHoldingsPathPrefix)
	VoucherTransfersURL, _ = url.JoinPath(dataURLBase, voucherTransfersPathPrefix)
	VoucherDataURL, _ = url.JoinPath(dataURLBase, voucherDataPathPrefix)
	CheckAliasURL, _ = url.JoinPath(dataURLBase, AliasPrefix)

	return nil
}
