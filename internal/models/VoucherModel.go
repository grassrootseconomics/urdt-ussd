package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Voucher struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID
	ContractAddress string
	TokenSymbol     string
	TokenDecimals   int
	Balance         float64
	ActiveVouchers  []ActiveVoucher `gorm:"foreignKey:VoucherID"`
	Transactions    []Transaction   `gorm:"foreignKey:VoucherID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
