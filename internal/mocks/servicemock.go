package mocks

import (
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount() (*models.AccountResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.AccountResponse), args.Error(1)
}

func (m *MockAccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.BalanceResponse), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	args := m.Called(trackingId)
	return args.Get(0).(*models.TrackStatusResponse), args.Error(1)
}