package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SessionID      string    `gorm:"unique"`
	PublicKey      string
	CustodialID    string
	TrackingID     string
	FirstName      string
	FamilyName     string
	Offerings      string
	Gender         string
	YOB            string
	Location       string
	Status         string
	AccountPIN     string
	TemporaryPIN   string
	ActiveVouchers []ActiveVoucher `gorm:"foreignKey:UserID"`
	Transactions   []Transaction   `gorm:"foreignKey:UserID"`
	Vouchers       []Voucher       `gorm:"foreignKey:UserID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
