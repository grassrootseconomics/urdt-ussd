package mocks

import (
	"context"

	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(ctx context.Context) (*api.OKResponse, error) {
	args := m.Called()
	return args.Get(0).(*api.OKResponse), args.Error(1)
}

func (m *MockAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.BalanceResponse), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error) {
	args := m.Called(trackingId)
	return args.Get(0).(*models.TrackStatusResponse), args.Error(1)
}

func (m *MockAccountService) TrackAccountStatus(ctx context.Context,publicKey string) (*api.OKResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*api.OKResponse), args.Error(1)
}


func (m *MockAccountService) FetchVouchers(ctx context.Context, publicKey string) (*models.VoucherHoldingResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.VoucherHoldingResponse), args.Error(1)
}
