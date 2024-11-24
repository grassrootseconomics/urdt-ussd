package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.grassecon.net/urdt/ussd/internal/storage"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

// TransferMetadata helps organize data fields
type TransferMetadata struct {
	Senders        string
	Recipients     string
	TransferValues string
	Addresses      string
	TxHashes       string
	Dates          string
	Symbols        string
	Decimals       string
}

// ProcessTransfers converts transfers into formatted strings
func ProcessTransfers(transfers []dataserviceapi.Last10TxResponse) TransferMetadata {
	var data TransferMetadata
	var senders, recipients, transferValues, addresses, txHashes, dates, symbols, decimals []string

	for _, t := range transfers {
		senders = append(senders, t.Sender)
		recipients = append(recipients, t.Recipient)

		// Scale down the amount
		scaledBalance := ScaleDownBalance(t.TransferValue, t.TokenDecimals)
		transferValues = append(transferValues, scaledBalance)

		addresses = append(addresses, t.ContractAddress)
		txHashes = append(txHashes, t.TxHash)
		dates = append(dates, fmt.Sprintf("%s", t.DateBlock))
		symbols = append(symbols, t.TokenSymbol)
		decimals = append(decimals, t.TokenDecimals)
	}

	data.Senders = strings.Join(senders, "\n")
	data.Recipients = strings.Join(recipients, "\n")
	data.TransferValues = strings.Join(transferValues, "\n")
	data.Addresses = strings.Join(addresses, "\n")
	data.TxHashes = strings.Join(txHashes, "\n")
	data.Dates = strings.Join(dates, "\n")
	data.Symbols = strings.Join(symbols, "\n")
	data.Decimals = strings.Join(decimals, "\n")

	return data
}

// GetTransferData retrieves and matches transfer data
// returns a formatted string of the full transaction/statement
func GetTransferData(ctx context.Context, db storage.PrefixDb, publicKey string, index int) (string, error) {
	keys := []string{"txfrom", "txto", "txval", "txaddr", "txhash", "txdate", "txsym"}
	data := make(map[string]string)

	for _, key := range keys {
		value, err := db.Get(ctx, []byte(key))
		if err != nil {
			return "", fmt.Errorf("failed to get %s: %v", key, err)
		}
		data[key] = string(value)
	}

	// Split the data
	senders := strings.Split(string(data["txfrom"]), "\n")
	recipients := strings.Split(string(data["txto"]), "\n")
	values := strings.Split(string(data["txval"]), "\n")
	addresses := strings.Split(string(data["txaddr"]), "\n")
	hashes := strings.Split(string(data["txhash"]), "\n")
	dates := strings.Split(string(data["txdate"]), "\n")
	syms := strings.Split(string(data["txsym"]), "\n")

	// Check if index is within range
	if index < 1 || index > len(senders) {
		return "", fmt.Errorf("transaction not found: index %d out of range", index)
	}

	// Adjust for 0-based indexing
	i := index - 1
	transactionType := "received"
	party := fmt.Sprintf("from: %s", strings.TrimSpace(senders[i]))
	if strings.TrimSpace(senders[i]) == publicKey {
		transactionType = "sent"
		party = fmt.Sprintf("to: %s", strings.TrimSpace(recipients[i]))
	}

	formattedDate := formatDate(strings.TrimSpace(dates[i]))

	// Build the full transaction detail
	detail := fmt.Sprintf(
		"%s %s %s\n%s\ncontract address: %s\ntxhash: %s\ndate: %s",
		transactionType,
		strings.TrimSpace(values[i]),
		strings.TrimSpace(syms[i]),
		party,
		strings.TrimSpace(addresses[i]),
		strings.TrimSpace(hashes[i]),
		formattedDate,
	)

	return detail, nil
}

// Helper function to format date in desired output
func formatDate(dateStr string) string {
	parsedDate, err := time.Parse("2006-01-02 15:04:05 -0700 MST", dateStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return ""
	}
	return parsedDate.Format("2006-01-02 03:04:05 PM")
}
