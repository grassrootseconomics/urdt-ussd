package ussd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"testing"
	"time"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	"git.grassecon.net/urdt/ussd/internal/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	testdataloader "github.com/peteole/testdata-loader"
	"github.com/stretchr/testify/require"
)

var (
	baseDir   = testdataloader.GetBasePath()
	flagsPath = path.Join(baseDir, "services", "registration", "pp.csv")
)

func TestNewHandlers(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	accountService := server.TestAccountService{}
	if err != nil {
		t.Logf(err.Error())
	}
	t.Run("Valid UserDataStore", func(t *testing.T) {
		mockStore := &mocks.MockUserDataStore{}
		handlers, err := NewHandlers(fm.parser, mockStore, &accountService)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if handlers == nil {
			t.Fatal("expected handlers to be non-nil")
		}
		if handlers.userdataStore == nil {
			t.Fatal("expected userdataStore to be set in handlers")
		}
	})

	// Test case for nil userdataStore
	t.Run("Nil UserDataStore", func(t *testing.T) {
		appFlags := &asm.FlagParser{}

		handlers, err := NewHandlers(appFlags, nil, &accountService)

		if err == nil {
			t.Fatal("expected an error, got none")
		}
		if handlers != nil {
			t.Fatal("expected handlers to be nil")
		}
		if err.Error() != "cannot create handler with nil userdata store" {
			t.Fatalf("expected specific error, got %v", err)
		}
	})
}

func TestCreateAccount(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	// Create required mocks
	flag_account_created, err := fm.GetFlag("flag_account_created")
	if err != nil {
		t.Logf(err.Error())
	}
	// Define session ID and mock data
	sessionId := "session123"
	notFoundErr := db.ErrNotFound{}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		serverResponse *api.OKResponse
		expectedResult resource.Result
	}{
		{
			name: "Test account creation success",
			serverResponse: &api.OKResponse{
				Ok:          true,
				Description: "Account creation successed",
				Result: map[string]any{
					"trackingId": "1234567890",
					"publicKey":  "1235QERYU",
				},
			},
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_account_created},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDataStore := new(mocks.MockUserDataStore)
			mockCreateAccountService := new(mocks.MockAccountService)

			// Create a Handlers instance with the mock data store
			h := &Handlers{
				userdataStore:  mockDataStore,
				accountService: mockCreateAccountService,
				flagManager:    fm.parser,
			}

			data := map[utils.DataTyp]string{
				utils.DATA_TRACKING_ID: tt.serverResponse.Result["trackingId"].(string),
				utils.DATA_PUBLIC_KEY:  tt.serverResponse.Result["publicKey"].(string),
			}

			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACCOUNT_CREATED).Return([]byte(""), notFoundErr)
			mockCreateAccountService.On("CreateAccount").Return(tt.serverResponse, nil)

			for key, value := range data {
				mockDataStore.On("WriteEntry", ctx, sessionId, key, []byte(value)).Return(nil)
			}

			// Call the method you want to test
			res, err := h.CreateAccount(ctx, "create_account", []byte("some-input"))

			// Assert that no errors occurred
			assert.NoError(t, err)

			// Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)
		})
	}
}

func TestWithPersister(t *testing.T) {
	// Test case: Setting a persister
	h := &Handlers{}
	p := &persist.Persister{}

	result := h.WithPersister(p)

	assert.Equal(t, p, h.pe, "The persister should be set correctly.")
	assert.Equal(t, h, result, "The returned handler should be the same instance.")
}

func TestWithPersister_PanicWhenAlreadySet(t *testing.T) {
	// Test case: Panic on multiple calls
	h := &Handlers{pe: &persist.Persister{}}
	require.Panics(t, func() {
		h.WithPersister(&persist.Persister{})
	}, "Should panic when trying to set a persister again.")
}

func TestSaveFirstname(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	firstName := "John"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_FIRST_NAME, []byte(firstName)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveFirstname(ctx, "save_firstname", []byte(firstName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveFamilyname(t *testing.T) {
	// Create a new instance of UserDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	familyName := "Doeee"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_FAMILY_NAME, []byte(familyName)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveFamilyname(ctx, "save_familyname", []byte(familyName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveTemporaryPin(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	mockStore := new(mocks.MockUserDataStore)
	if err != nil {
		log.Fatal(err)
	}
	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

	// Create the Handlers instance with the mock flag manager
	h := &Handlers{
		flagManager:   fm.parser,
		userdataStore: mockStore,
	}
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Valid Pin entry",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
			},
		},
		{
			name:  "Invalid Pin entry",
			input: []byte("12343"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_pin},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up the expected behavior of the mock
			mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_PIN, []byte(tt.input)).Return(nil)

			// Call the method
			res, err := h.SaveTemporaryPin(ctx, "save_pin", tt.input)

			if err != nil {
				t.Error(err)
			}

			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should be equal to account created")

		})
	}
}

func TestSaveYoB(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	yob := "1980"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_YOB, []byte(yob)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveYob(ctx, "save_yob", []byte(yob))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveLocation(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	yob := "Kilifi"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_LOCATION, []byte(yob)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveLocation(ctx, "save_location", []byte(yob))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveOfferings(t *testing.T) {
	// Create a new instance of MockUserDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	offerings := "Bananas"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_OFFERINGS, []byte(offerings)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveOfferings(ctx, "save_offerings", []byte(offerings))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveGender(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)
	mockState := state.NewState(16)

	// Define the session ID and context
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name            string
		input           []byte
		expectedGender  string
		expectCall      bool
		executingSymbol string
	}{
		{
			name:            "Valid Male Input",
			input:           []byte("1"),
			expectedGender:  "male",
			executingSymbol: "set_male",
			expectCall:      true,
		},
		{
			name:            "Valid Female Input",
			input:           []byte("2"),
			expectedGender:  "female",
			executingSymbol: "set_female",
			expectCall:      true,
		},
		{
			name:            "Valid Unspecified Input",
			input:           []byte("3"),
			executingSymbol: "set_unspecified",
			expectedGender:  "unspecified",
			expectCall:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations for the mock database
			if tt.expectCall {
				expectedKey := utils.DATA_GENDER
				mockStore.On("WriteEntry", ctx, sessionId, expectedKey, []byte(tt.expectedGender)).Return(nil)
			} else {
				mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender)).Return(nil)
			}
			mockState.ExecPath = append(mockState.ExecPath, tt.executingSymbol)
			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: mockStore,
				st:            mockState,
			}

			// Call the method
			_, err := h.SaveGender(ctx, "save_gender", tt.input)

			// Assert no error
			assert.NoError(t, err)

			// Verify expectations
			if tt.expectCall {
				mockStore.AssertCalled(t, "WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender))
			} else {
				mockStore.AssertNotCalled(t, "WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender))
			}
		})
	}
}

func TestCheckIdentifier(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define the session ID and context
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name            string
		mockPublicKey   []byte
		mockErr         error
		expectedContent string
		expectError     bool
	}{
		{
			name:            "Saved public Key",
			mockPublicKey:   []byte("0xa8363"),
			mockErr:         nil,
			expectedContent: "0xa8363",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations for the mock database
			mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.mockPublicKey, tt.mockErr)

			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: mockStore,
			}

			// Call the method
			res, err := h.CheckIdentifier(ctx, "check_identifier", nil)

			// Assert results

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, res.Content)

			// Verify expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSender(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "254712345678"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, _ := h.GetSender(ctx, "get_sender", []byte(""))

	//Assert that the sessionId is what was set as the result content.
	assert.Equal(t, sessionId, res.Content)
}

func TestGetAmount(t *testing.T) {
	mockDataStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	amount := "0.03"
	activeSym := "SRF"

	// Set up the expected behavior of the mock
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM).Return([]byte(activeSym), nil)
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_AMOUNT).Return([]byte(amount), nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockDataStore,
	}

	// Call the method
	res, _ := h.GetAmount(ctx, "get_amount", []byte(""))

	formattedAmount := fmt.Sprintf("%s %s", amount, activeSym)

	//Assert that the retrieved amount is what was set as the content
	assert.Equal(t, formattedAmount, res.Content)

}

func TestGetRecipient(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	recepient := "0xcasgatweksalw1018221"

	// Set up the expected behavior of the mock
	mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_RECIPIENT).Return([]byte(recepient), nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, _ := h.GetRecipient(ctx, "get_recipient", []byte("Getting recipient..."))

	//Assert that the retrieved recepient is what was set as the content
	assert.Equal(t, recepient, res.Content)

}

func TestGetFlag(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	expectedFlag := uint32(9)
	if err != nil {
		t.Logf(err.Error())
	}
	flag, err := fm.GetFlag("flag_account_created")
	if err != nil {
		t.Logf(err.Error())
	}

	assert.Equal(t, uint32(flag), expectedFlag, "Flags should be equal to account created")
}

func TestSetLanguage(t *testing.T) {
	// Create a new instance of the Flag Manager
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		execPath       []string
		expectedResult resource.Result
	}{
		{
			name:     "Set Default Language (English)",
			execPath: []string{"set_eng"},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "eng",
			},
		},
		{
			name:     "Set Swahili Language",
			execPath: []string{"set_swa"},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "swa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockState := state.NewState(16)
			// Set the ExecPath
			mockState.ExecPath = tt.execPath

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
				st:          mockState,
			}

			// Call the method
			res, err := h.SetLanguage(context.Background(), "set_language", nil)

			if err != nil {
				t.Error(err)
			}

			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should match expected result")

		})
	}
}
func TestResetAllowUpdate(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.parser.GetFlag("flag_allow_update")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Resets allow update",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_allow_update},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetAllowUpdate(context.Background(), "reset_allow update", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should be equal to account created")

		})
	}
}

func TestResetAccountAuthorized(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Resets account authorized",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetAccountAuthorized(context.Background(), "reset_account_authorized", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestIncorrectPinReset(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test incorrect pin reset",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetIncorrectPin(context.Background(), "reset_incorrect_pin", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestResetIncorrectYob(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test incorrect yob reset",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_date_format},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetIncorrectYob(context.Background(), "reset_incorrect_yob", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestAuthorize(t *testing.T) {

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	//expectedResult := resource.Result{}
	mockState := state.NewState(16)
	flag_incorrect_pin, _ := fm.GetFlag("flag_incorrect_pin")
	flag_account_authorized, _ := fm.GetFlag("flag_account_authorized")
	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	//Assuming 1234 is the correct account pin
	accountPIN := "1234"

	// Define session ID and mock data
	sessionId := "session123"
	typ := utils.DATA_ACCOUNT_PIN

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct pin",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
				FlagSet:   []uint32{flag_allow_update, flag_account_authorized},
			},
		},
		{
			name:  "Test with incorrect pin",
			input: []byte("1235"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
				FlagSet:   []uint32{flag_incorrect_pin},
			},
		},
		{
			name:           "Test with pin that is not a 4 digit",
			input:          []byte("1235aqds"),
			expectedResult: resource.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create context with session ID
			ctx := context.WithValue(context.Background(), "SessionId", sessionId)

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte(accountPIN), nil)

			// Create a Handlers instance with the mock data store

			// Call the method under test
			res, err := h.Authorize(ctx, "authorize", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestVerifyYob(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	sessionId := "session123"
	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)
	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct yob",
			input: []byte("1980"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_date_format},
			},
		},
		{
			name:  "Test with incorrect yob",
			input: []byte("sgahaha"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_date_format},
			},
		},
		{
			name:  "Test with numeric but less 4 digits",
			input: []byte("123"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_date_format},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Call the method under test
			res, err := h.VerifyYob(ctx, "verify_yob", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestVerifyCreatePin(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	sessionId := "session123"
	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	flag_pin_set, _ := fm.parser.GetFlag("flag_pin_set")
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	//Assuming this was the first set PIN to verify against
	firstSetPin := "1234"

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct PIN confirmation",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_valid_pin, flag_pin_set},
				FlagReset: []uint32{flag_pin_mismatch},
			},
		},
		{
			name:  "Test with PIN that does not match first ",
			input: []byte("1324"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_pin_mismatch},
			},
		},
	}

	typ := utils.DATA_TEMPORARY_PIN

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte(firstSetPin), nil)

			// Set up the expected behavior of the mock
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_PIN, []byte(firstSetPin)).Return(nil)

			// Call the method under test
			res, err := h.VerifyCreatePin(ctx, "verify_create_pin", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestCheckAccountStatus(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	sessionId := "session123"
	flag_account_success, _ := fm.GetFlag("flag_account_success")
	flag_account_pending, _ := fm.GetFlag("flag_account_pending")
	flag_api_error, _ := fm.GetFlag("flag_api_call_error")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		input          []byte
		serverResponse *api.OKResponse
		response       *models.TrackStatusResponse
		expectedResult resource.Result
	}{
		{
			name:  "Test when account is on the Sarafu network",
			input: []byte("TrackingId1234"),
			serverResponse: &api.OKResponse{
				Ok:          true,
				Description: "Account creation succeeded",
				Result: map[string]any{
					"active": true,
				},
			},
			response: &models.TrackStatusResponse{
				Ok: true,
				Result: struct {
					Transaction struct {
						CreatedAt     time.Time   "json:\"createdAt\""
						Status        string      "json:\"status\""
						TransferValue json.Number "json:\"transferValue\""
						TxHash        string      "json:\"txHash\""
						TxType        string      "json:\"txType\""
					}
				}{
					Transaction: models.Transaction{
						CreatedAt:     time.Now(),
						Status:        "SUCCESS",
						TransferValue: json.Number("0.5"),
						TxHash:        "0x123abc456def",
						TxType:        "transfer",
					},
				},
			},
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_account_success},
				FlagReset: []uint32{flag_api_error, flag_account_pending},
			},
		},
		{
			name:  "Test when the account is not  yet on the sarafu network",
			input: []byte("TrackingId1234"),
			response: &models.TrackStatusResponse{
				Ok: true,
				Result: struct {
					Transaction struct {
						CreatedAt     time.Time   "json:\"createdAt\""
						Status        string      "json:\"status\""
						TransferValue json.Number "json:\"transferValue\""
						TxHash        string      "json:\"txHash\""
						TxType        string      "json:\"txType\""
					}
				}{
					Transaction: models.Transaction{
						CreatedAt:     time.Now(),
						Status:        "SUCCESS",
						TransferValue: json.Number("0.5"),
						TxHash:        "0x123abc456def",
						TxType:        "transfer",
					},
				},
			},
			serverResponse: &api.OKResponse{
				Ok:          true,
				Description: "Account creation succeeded",
				Result: map[string]any{
					"active": false,
				},
			},
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_account_pending},
				FlagReset: []uint32{flag_api_error, flag_account_success},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore := new(mocks.MockUserDataStore)
			mockCreateAccountService := new(mocks.MockAccountService)

			h := &Handlers{
				userdataStore:  mockDataStore,
				accountService: mockCreateAccountService,
				flagManager:    fm.parser,
			}

			status := tt.response.Result.Transaction.Status
			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.input, nil)

			mockCreateAccountService.On("CheckAccountStatus", string(tt.input)).Return(tt.response, nil)
			mockCreateAccountService.On("TrackAccountStatus", string(tt.input)).Return(tt.serverResponse, nil)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_STATUS, []byte(status)).Return(nil).Maybe()

			// Call the method under test
			res, _ := h.CheckAccountStatus(ctx, "check_account_status", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestTransactionReset(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_invalid_recipient, _ := fm.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := fm.GetFlag("flag_invalid_recipient_with_invite")

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test transaction reset for amount and recipient",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_invalid_recipient, flag_invalid_recipient_with_invite},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, []byte("")).Return(nil)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_RECIPIENT, []byte("")).Return(nil)

			// Call the method under test
			res, _ := h.TransactionReset(ctx, "transaction_reset", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestResetInvalidAmount(t *testing.T) {
	sessionId := "session123"

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test amount reset",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_invalid_amount},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, []byte("")).Return(nil)

			// Call the method under test
			res, _ := h.ResetTransactionAmount(ctx, "transaction_reset_amount", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestInitiateTransaction(t *testing.T) {
	sessionId := "254712345678"

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	account_authorized_flag, err := fm.parser.GetFlag("flag_account_authorized")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name           string
		input          []byte
		Recipient      []byte
		Amount         []byte
		ActiveSym      []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name:      "Test initiate transaction",
			Amount:    []byte("0.002"),
			ActiveSym: []byte("SRF"),
			Recipient: []byte("0x12415ass27192"),
			expectedResult: resource.Result{
				FlagReset: []uint32{account_authorized_flag},
				Content:   "Your request has been sent. 0x12415ass27192 will receive 0.002 SRF from 254712345678.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_AMOUNT).Return(tt.Amount, nil)
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_RECIPIENT).Return(tt.Recipient, nil)
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM).Return(tt.ActiveSym, nil)

			// Call the method under test
			res, _ := h.InitiateTransaction(ctx, "transaction_reset_amount", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)
		})
	}
}

func TestQuit(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test quit message",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
				Content:   "Thank you for using Sarafu. Goodbye!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Call the method under test
			res, _ := h.Quit(ctx, "test_quit", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}
func TestIsValidPIN(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		expected bool
	}{
		{
			name:     "Valid PIN with 4 digits",
			pin:      "1234",
			expected: true,
		},
		{
			name:     "Valid PIN with leading zeros",
			pin:      "0001",
			expected: true,
		},
		{
			name:     "Invalid PIN with less than 4 digits",
			pin:      "123",
			expected: false,
		},
		{
			name:     "Invalid PIN with more than 4 digits",
			pin:      "12345",
			expected: false,
		},
		{
			name:     "Invalid PIN with letters",
			pin:      "abcd",
			expected: false,
		},
		{
			name:     "Invalid PIN with special characters",
			pin:      "12@#",
			expected: false,
		},
		{
			name:     "Empty PIN",
			pin:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidPIN(tt.pin)
			if actual != tt.expected {
				t.Errorf("isValidPIN(%q) = %v; expected %v", tt.pin, actual, tt.expected)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		activeBal      []byte
		balance        string
		expectedResult resource.Result
	}{
		{
			name:      "Test with valid amount",
			input:     []byte("4.10"),
			activeBal: []byte("5"),
			expectedResult: resource.Result{
				Content: "4.10",
			},
		},
		{
			name:      "Test with amount larger than active balance",
			input:     []byte("5.02"),
			activeBal: []byte("5"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_amount},
				Content: "5.02",
			},
		},
		{
			name:    "Test with invalid amount format",
			input:   []byte("0.02ms"),
			activeBal: []byte("5"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_amount},
				Content: "0.02ms",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock behavior for active balance retrieval
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_BAL).Return(tt.activeBal, nil)

			// Mock behavior for storing the amount (if valid)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, tt.input).Return(nil).Maybe()

			// Call the method under test
			res, _ := h.ValidateAmount(ctx, "test_validate_amount", tt.input)

			// Assert no errors occurred
			assert.NoError(t, err)

			// Assert the result matches the expected result
			assert.Equal(t, tt.expectedResult, res, "Expected result should match actual result")

			// Assert all expectations were met
			mockDataStore.AssertExpectations(t)
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_invalid_recipient, _ := fm.parser.GetFlag("flag_invalid_recipient")
	mockDataStore := new(mocks.MockUserDataStore)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with invalid recepient",
			input: []byte("000"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_recipient},
				Content: "000",
			},
		},
		{
			name:           "Test with valid recepient",
			input:          []byte("0705X2"),
			expectedResult: resource.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_RECIPIENT, tt.input).Return(nil)

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager:   fm.parser,
				userdataStore: mockDataStore,
			}

			// Call the method
			res, err := h.ValidateRecipient(ctx, "validate_recepient", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestCheckBalance(t *testing.T) {
	tests := []struct {
		name           string
		sessionId      string
		publicKey      string
		activeSym      string
		activeBal      string
		expectedResult resource.Result
		expectError    bool
	}{
		{
			name:           "User with active sym",
			sessionId:      "session456",
			publicKey:      "0X98765432109",
			activeSym:      "ETH",
			activeBal:      "1.5",
			expectedResult: resource.Result{Content: "Balance: 1.5 ETH\n"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore := new(mocks.MockUserDataStore)
			mockAccountService := new(mocks.MockAccountService)
			ctx := context.WithValue(context.Background(), "SessionId", tt.sessionId)

			h := &Handlers{
				userdataStore:  mockDataStore,
				accountService: mockAccountService,
			}

			// Mock for user with active sym
			mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_ACTIVE_SYM).Return([]byte(tt.activeSym), nil)
			mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_ACTIVE_BAL).Return([]byte(tt.activeBal), nil)

			res, err := h.CheckBalance(ctx, "check_balance", []byte(""))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res, "Result should match expected output")
			}

			mockDataStore.AssertExpectations(t)
			mockAccountService.AssertExpectations(t)
		})
	}
}

func TestGetProfile(t *testing.T) {

	sessionId := "session123"

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		st:             mockState,
	}

	tests := []struct {
		name         string
		languageCode string
		keys         []utils.DataTyp
		profileInfo  []string
		result       resource.Result
	}{
		{
			name:         "Test with full profile information in eng",
			keys:         []utils.DataTyp{utils.DATA_FAMILY_NAME, utils.DATA_FIRST_NAME, utils.DATA_GENDER, utils.DATA_OFFERINGS, utils.DATA_LOCATION, utils.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "eng",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
					"John Doee", "Male", "48", "Kilifi", "Bananas",
				),
			},
		},
		{
			name:         "Test with with profile information in swa ",
			keys:         []utils.DataTyp{utils.DATA_FAMILY_NAME, utils.DATA_FIRST_NAME, utils.DATA_GENDER, utils.DATA_OFFERINGS, utils.DATA_LOCATION, utils.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "swa",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Jina: %s\nJinsia: %s\nUmri: %s\nEneo: %s\nUnauza: %s\n",
					"John Doee", "Male", "48", "Kilifi", "Bananas",
				),
			},
		},
		{
			name:         "Test with with profile information with language that is not yet supported",
			keys:         []utils.DataTyp{utils.DATA_FAMILY_NAME, utils.DATA_FIRST_NAME, utils.DATA_GENDER, utils.DATA_OFFERINGS, utils.DATA_LOCATION, utils.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "nor",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
					"John Doee", "Male", "48", "Kilifi", "Bananas",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "SessionId", sessionId)
			ctx = context.WithValue(ctx, "Language", lang.Language{
				Code: tt.languageCode,
			})
			for index, key := range tt.keys {
				mockDataStore.On("ReadEntry", ctx, sessionId, key).Return([]byte(tt.profileInfo[index]), nil).Maybe()
			}

			res, _ := h.GetProfileInfo(ctx, "get_profile_info", []byte(""))

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.result, "Result should contain profile information served back to user")

		})
	}
}

func TestVerifyNewPin(t *testing.T) {
	sessionId := "session123"

	fm, _ := NewFlagManager(flagsPath)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		userdataStore:  mockDataStore,
		flagManager:    fm.parser,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with valid pin",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_valid_pin},
			},
		},
		{
			name:  "Test with invalid pin",
			input: []byte("123"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_valid_pin},
			},
		},
		{
			name:  "Test with invalid pin",
			input: []byte("12345"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_valid_pin},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Call the function under test
			res, _ := h.VerifyNewPin(ctx, "verify_new_pin", tt.input)

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}

}

func TestConfirmPin(t *testing.T) {
	sessionId := "session123"

	fm, _ := NewFlagManager(flagsPath)
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		userdataStore:  mockDataStore,
		flagManager:    fm.parser,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		input          []byte
		temporarypin   []byte
		expectedResult resource.Result
	}{
		{
			name:         "Test with correct pin confirmation",
			input:        []byte("1234"),
			temporarypin: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_pin_mismatch},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the expected behavior of the mock
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_PIN, []byte(tt.temporarypin)).Return(nil)

			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_PIN).Return(tt.temporarypin, nil)

			//Call the function under test
			res, _ := h.ConfirmPinChange(ctx, "confirm_pin_change", tt.temporarypin)

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}
}

func TestSetVoucher(t *testing.T) {
	mockDataStore := new(mocks.MockUserDataStore)

	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	temporarySym := []byte("tempSym")
	temporaryBal := []byte("tempBal")

	// Set expectations for the mock data store
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_SYM).Return(temporarySym, nil)
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_BAL).Return(temporaryBal, nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM, temporarySym).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACTIVE_BAL, temporaryBal).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_SYM, []byte("")).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_BAL, []byte("")).Return(nil)

	h := &Handlers{
		userdataStore: mockDataStore,
	}

	// Call the method under test
	res, err := h.SetVoucher(ctx, "someSym", []byte{})

	// Assert that no errors occurred
	assert.NoError(t, err)

	// Assert that the result content is correct
	assert.Equal(t, string(temporarySym), res.Content)

	// Assert that expectations were met
	mockDataStore.AssertExpectations(t)
}

func TestFetchCustodialBalances(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	flag_api_error, _ := fm.GetFlag("flag_api_call_error")

	// Define test data
	sessionId := "session123"
	publicKey := "0X13242618721"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		balanceResonse *models.BalanceResponse
		expectedResult resource.Result
	}{
		{
			name: "Test when fetch custodial balances is not a success",
			balanceResonse: &models.BalanceResponse{
				Ok: false,
				Result: struct {
					Balance string      `json:"balance"`
					Nonce   json.Number `json:"nonce"`
				}{
					Balance: "0.003 CELO",
					Nonce:   json.Number("0"),
				},
			},
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_api_error},
			},
		},
		{
			name: "Test when fetch custodial balances is a success",
			balanceResonse: &models.BalanceResponse{
				Ok: true,
				Result: struct {
					Balance string      `json:"balance"`
					Nonce   json.Number `json:"nonce"`
				}{
					Balance: "0.003 CELO",
					Nonce:   json.Number("0"),
				},
			},
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_api_error},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDataStore := new(mocks.MockUserDataStore)
			mockCreateAccountService := new(mocks.MockAccountService)
			mockState := state.NewState(16)

			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore:  mockDataStore,
				flagManager:    fm.parser,
				st:             mockState,
				accountService: mockCreateAccountService,
			}

			// Set up the expected behavior of the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return([]byte(publicKey), nil)
			mockCreateAccountService.On("CheckBalance", string(publicKey)).Return(tt.balanceResonse, nil)

			// Call the method
			res, _ := h.FetchCustodialBalances(ctx, "fetch_custodial_balances", []byte(""))

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}
}
