package config

import (
	"net/url"
	"strings"

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
	defaultLanguage		   = "eng"
	languages []string
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
	DbConn		string
	DefaultLanguage	    string
	Languages	[]string
)

func setLanguage() error {
	defaultLanguage = initializers.GetEnv("DEFAULT_LANGUAGE", defaultLanguage)
	languages = strings.Split(initializers.GetEnv("LANGUAGES", defaultLanguage), ",")
	haveDefaultLanguage := false
	for i, v := range(languages) {
		languages[i] = strings.ReplaceAll(v, " ", "")
		if languages[i] == defaultLanguage {
			haveDefaultLanguage = true
		}
	}

	if !haveDefaultLanguage {
		languages = append([]string{defaultLanguage}, languages...)
	}

	return nil
}

func setBase() error {
	var err error

	custodialURLBase = initializers.GetEnv("CUSTODIAL_URL_BASE", "http://localhost:5003")
	dataURLBase = initializers.GetEnv("DATA_URL_BASE", "http://localhost:5006")
	BearerToken = initializers.GetEnv("BEARER_TOKEN", "")

	_, err = url.Parse(custodialURLBase)
	if err != nil {
		return err
	}
	_, err = url.Parse(dataURLBase)
	if err != nil {
		return err
	}

	return nil
}

func setConn() error {
	DbConn = initializers.GetEnv("DB_CONN", "")
	return nil
}

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() error {
	err := setBase()
	if err != nil {
		return err
	}
	err = setConn()
	if err != nil {
		return err
	}
	err = setLanguage()
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
	DefaultLanguage = defaultLanguage
	Languages = languages

	return nil
}
