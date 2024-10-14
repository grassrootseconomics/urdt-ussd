package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type ActiveVoucher struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID
	VoucherID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
