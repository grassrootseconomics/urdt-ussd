package common

import (
	"context"
	"fmt"
	"strings"

	"git.grassecon.net/urdt/ussd/internal/storage"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

// VoucherMetadata helps organize data fields
type VoucherMetadata struct {
	Symbols   string
	Balances  string
	Decimals  string
	Addresses string
}

// ProcessVouchers converts holdings into formatted strings
func ProcessVouchers(holdings []dataserviceapi.TokenHoldings) VoucherMetadata {
	var data VoucherMetadata
	var symbols, balances, decimals, addresses []string

	for i, h := range holdings {
		symbols = append(symbols, fmt.Sprintf("%d:%s", i+1, h.TokenSymbol))
		balances = append(balances, fmt.Sprintf("%d:%s", i+1, h.Balance))
		decimals = append(decimals, fmt.Sprintf("%d:%s", i+1, h.TokenDecimals))
		addresses = append(addresses, fmt.Sprintf("%d:%s", i+1, h.ContractAddress))
	}

	data.Symbols = strings.Join(symbols, "\n")
	data.Balances = strings.Join(balances, "\n")
	data.Decimals = strings.Join(decimals, "\n")
	data.Addresses = strings.Join(addresses, "\n")

	return data
}

//func StoreVouchers(db storage.PrefixDb, data VoucherMetadata) {
//		value, err := db.Put(ctx, []byte(key))
//		if err != nil {
//			return nil, fmt.Errorf("failed to get %s: %v", key, err)
//		}
//		data[key] = string(value)
//	}
//}

// GetVoucherData retrieves and matches voucher data
func GetVoucherData(ctx context.Context, db storage.PrefixDb, input string) (*dataserviceapi.TokenHoldings, error) {
	keys := []string{"sym", "bal", "deci", "addr"}
	data := make(map[string]string)

	for _, key := range keys {
		value, err := db.Get(ctx, []byte(key))
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %v", key, err)
		}
		data[key] = string(value)
	}

	symbol, balance, decimal, address := MatchVoucher(input,
		data["sym"],
		data["bal"],
		data["deci"],
		data["addr"])

	if symbol == "" {
		return nil, nil
	}

	return &dataserviceapi.TokenHoldings{
		TokenSymbol:     string(symbol),
		Balance:         string(balance),
		TokenDecimals:   string(decimal),
		ContractAddress: string(address),
	}, nil
}

// MatchVoucher finds the matching voucher symbol, balance, decimals and contract address based on the input.
func MatchVoucher(input, symbols, balances, decimals, addresses string) (symbol, balance, decimal, address string) {
	symList := strings.Split(symbols, "\n")
	balList := strings.Split(balances, "\n")
	decList := strings.Split(decimals, "\n")
	addrList := strings.Split(addresses, "\n")

	logg.Tracef("found" , "symlist", symList, "syms", symbols, "input", input)
	for i, sym := range symList {
		parts := strings.SplitN(sym, ":", 2)

		if input == parts[0] || strings.EqualFold(input, parts[1]) {
			symbol = parts[1]
			if i < len(balList) {
				balance = strings.SplitN(balList[i], ":", 2)[1]
			}
			if i < len(decList) {
				decimal = strings.SplitN(decList[i], ":", 2)[1]
			}
			if i < len(addrList) {
				address = strings.SplitN(addrList[i], ":", 2)[1]
			}
			break
		}
	}
	return
}

// StoreTemporaryVoucher saves voucher metadata as temporary entries in the DataStore.
func StoreTemporaryVoucher(ctx context.Context, store DataStore, sessionId string, data *dataserviceapi.TokenHoldings) error {
	tempData := fmt.Sprintf("%s,%s,%s,%s", data.TokenSymbol, data.Balance, data.TokenDecimals, data.ContractAddress)

	if err := store.WriteEntry(ctx, sessionId, DATA_TEMPORARY_VALUE, []byte(tempData)); err != nil {
		return err
	}

	return nil
}

// GetTemporaryVoucherData retrieves temporary voucher metadata from the DataStore.
func GetTemporaryVoucherData(ctx context.Context, store DataStore, sessionId string) (*dataserviceapi.TokenHoldings, error) {
	temp_data, err := store.ReadEntry(ctx, sessionId, DATA_TEMPORARY_VALUE)
	if err != nil {
		return nil, err
	}

	values := strings.SplitN(string(temp_data), ",", 4)

	data := &dataserviceapi.TokenHoldings{}

	data.TokenSymbol = values[0]
	data.Balance = values[1]
	data.TokenDecimals = values[2]
	data.ContractAddress = values[3]

	return data, nil
}

// UpdateVoucherData sets the active voucher data and clears the temporary voucher data in the DataStore.
func UpdateVoucherData(ctx context.Context, store DataStore, sessionId string, data *dataserviceapi.TokenHoldings) error {
	logg.TraceCtxf(ctx, "dtal", "data", data)
	// Active voucher data entries
	activeEntries := map[DataTyp][]byte{
		DATA_ACTIVE_SYM:     []byte(data.TokenSymbol),
		DATA_ACTIVE_BAL:     []byte(data.Balance),
		DATA_ACTIVE_DECIMAL: []byte(data.TokenDecimals),
		DATA_ACTIVE_ADDRESS: []byte(data.ContractAddress),
	}

	// Write active data
	for key, value := range activeEntries {
		if err := store.WriteEntry(ctx, sessionId, key, value); err != nil {
			return err
		}
	}

	return nil
}
