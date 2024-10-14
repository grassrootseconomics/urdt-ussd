package initializers

import "git.grassecon.net/urdt/ussd/internal/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.ActiveVoucher{})
	DB.AutoMigrate(&models.Voucher{})
	DB.AutoMigrate(&models.Transaction{})
}
