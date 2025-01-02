package ussd

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/internal/testutil/mocks"
	"git.grassecon.net/urdt/ussd/internal/testutil/testservice"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"git.grassecon.net/urdt/ussd/models"

	"git.grassecon.net/urdt/ussd/common"
	"github.com/alecthomas/assert/v2"

	testdataloader "github.com/peteole/testdata-loader"
	"github.com/stretchr/testify/require"

	visedb "git.defalsify.org/vise.git/db"
	memdb "git.defalsify.org/vise.git/db/mem"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

var (
	baseDir   = testdataloader.GetBasePath()
	flagsPath = path.Join(baseDir, "services", "registration", "pp.csv")
)

// mockReplaceSeparator function
var mockReplaceSeparator = func(input string) string {
	return strings.ReplaceAll(input, ":", ": ")
}

// InitializeTestStore sets up and returns an in-memory database and store.
func InitializeTestStore(t *testing.T) (context.Context, *common.UserDataStore) {
	ctx := context.Background()

	// Initialize memDb
	db := memdb.NewMemDb()
	err := db.Connect(ctx, "")
	require.NoError(t, err, "Failed to connect to memDb")

	// Create UserDataStore with memDb
	store := &common.UserDataStore{Db: db}

	t.Cleanup(func() {
		db.Close() // Ensure the DB is closed after each test
	})

	return ctx, store
}

func InitializeTestSubPrefixDb(t *testing.T, ctx context.Context) *storage.SubPrefixDb {
	db := memdb.NewMemDb()
	err := db.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	prefix := common.ToBytes(visedb.DATATYPE_USERDATA)
	spdb := storage.NewSubPrefixDb(db, prefix)

	return spdb
}

func TestNewHandlers(t *testing.T) {
	_, store := InitializeTestStore(t)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		log.Fatal(err)
	}

	accountService := testservice.TestAccountService{}

	// Test case for valid UserDataStore
	t.Run("Valid UserDataStore", func(t *testing.T) {
		handlers, err := NewHandlers(fm.parser, store, nil, &accountService, mockReplaceSeparator)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if handlers == nil {
			t.Fatal("expected handlers to be non-nil")
		}
		if handlers.userdataStore == nil {
			t.Fatal("expected userdataStore to be set in handlers")
		}
		if handlers.ReplaceSeparatorFunc == nil {
			t.Fatal("expected ReplaceSeparatorFunc to be set in handlers")
		}

		// Test ReplaceSeparatorFunc functionality
		input := "1:Menu item"
		expectedOutput := "1: Menu item"
		if handlers.ReplaceSeparatorFunc(input) != expectedOutput {
			t.Fatalf("ReplaceSeparatorFunc function did not return expected output: got %v, want %v", handlers.ReplaceSeparatorFunc(input), expectedOutput)
		}
	})

	// Test case for nil UserDataStore
	t.Run("Nil UserDataStore", func(t *testing.T) {
		handlers, err := NewHandlers(fm.parser, nil, nil, &accountService, mockReplaceSeparator)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
		if handlers != nil {
			t.Fatal("expected handlers to be nil")
		}
		expectedError := "cannot create handler with nil userdata store"
		if err.Error() != expectedError {
			t.Fatalf("expected error '%s', got '%v'", expectedError, err)
		}
	})
}

func TestInit(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	adminstore, err := utils.NewAdminStore(ctx, "admin_numbers")
	if err != nil {
		t.Fatal(err.Error())
	}

	st := state.NewState(128)
	ca := cache.NewCache()

	flag_admin_privilege, _ := fm.GetFlag("flag_admin_privilege")

	tests := []struct {
		name           string
		setup          func() (*Handlers, context.Context)
		input          []byte
		expectedResult resource.Result
	}{
		{
			name: "Handler not ready",
			setup: func() (*Handlers, context.Context) {
				return &Handlers{}, ctx
			},
			input:          []byte("1"),
			expectedResult: resource.Result{},
		},
		{
			name: "State and memory initialization",
			setup: func() (*Handlers, context.Context) {
				pe := persist.NewPersister(store).WithSession(sessionId).WithContent(st, ca)
				h := &Handlers{
					flagManager: fm.parser,
					adminstore:  adminstore,
					pe:          pe,
				}
				return h, context.WithValue(ctx, "SessionId", sessionId)
			},
			input: []byte("1"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_admin_privilege},
			},
		},
		{
			name: "Non-admin session initialization",
			setup: func() (*Handlers, context.Context) {
				pe := persist.NewPersister(store).WithSession("0712345678").WithContent(st, ca)
				h := &Handlers{
					flagManager: fm.parser,
					adminstore:  adminstore,
					pe:          pe,
				}
				return h, context.WithValue(context.Background(), "SessionId", "0712345678")
			},
			input: []byte("1"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_admin_privilege},
			},
		},
		{
			name: "Move to top node on empty input",
			setup: func() (*Handlers, context.Context) {
				pe := persist.NewPersister(store).WithSession(sessionId).WithContent(st, ca)
				h := &Handlers{
					flagManager: fm.parser,
					adminstore:  adminstore,
					pe:          pe,
				}
				st.Code = []byte("some pending bytecode")
				return h, context.WithValue(ctx, "SessionId", sessionId)
			},
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_admin_privilege},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, testCtx := tt.setup()
			res, err := h.Init(testCtx, "", tt.input)

			assert.NoError(t, err, "Unexpected error occurred")
			assert.Equal(t, res, tt.expectedResult, "Expected result should match actual result")
		})
	}
}

func TestCreateAccount(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}

	flag_account_created, err := fm.GetFlag("flag_account_created")
	if err != nil {
		t.Logf(err.Error())
	}

	tests := []struct {
		name           string
		serverResponse *models.AccountResult
		expectedResult resource.Result
	}{
		{
			name: "Test account creation success",
			serverResponse: &models.AccountResult{
				TrackingId: "1234567890",
				PublicKey:  "0xD3adB33f",
			},
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_account_created},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccountService := new(mocks.MockAccountService)

			h := &Handlers{
				userdataStore:  store,
				accountService: mockAccountService,
				flagManager:    fm.parser,
			}

			mockAccountService.On("CreateAccount").Return(tt.serverResponse, nil)

			// Call the method you want to test
			res, err := h.CreateAccount(ctx, "create_account", []byte(""))

			// Assert that no errors occurred
			assert.NoError(t, err)

			// Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
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
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_firstname_set, _ := fm.GetFlag("flag_firstname_set")

	// Set the flag in the State
	mockState := state.NewState(128)
	mockState.SetFlag(flag_allow_update)

	expectedResult := resource.Result{}

	// Define test data
	firstName := "John"

	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(firstName)); err != nil {
		t.Fatal(err)
	}

	expectedResult.FlagSet = []uint32{flag_firstname_set}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
		flagManager:   fm.parser,
		st:            mockState,
	}

	// Call the method
	res, err := h.SaveFirstname(ctx, "save_firstname", []byte(firstName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	// Verify that the DATA_FIRST_NAME entry has been updated with the temporary value
	storedFirstName, _ := store.ReadEntry(ctx, sessionId, common.DATA_FIRST_NAME)
	assert.Equal(t, firstName, string(storedFirstName))
}

func TestSaveFamilyname(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_firstname_set, _ := fm.GetFlag("flag_familyname_set")

	// Set the flag in the State
	mockState := state.NewState(128)
	mockState.SetFlag(flag_allow_update)

	expectedResult := resource.Result{}

	expectedResult.FlagSet = []uint32{flag_firstname_set}

	// Define test data
	familyName := "Doeee"

	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(familyName)); err != nil {
		t.Fatal(err)
	}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
		st:            mockState,
		flagManager:   fm.parser,
	}

	// Call the method
	res, err := h.SaveFamilyname(ctx, "save_familyname", []byte(familyName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	// Verify that the DATA_FAMILY_NAME entry has been updated with the temporary value
	storedFamilyName, _ := store.ReadEntry(ctx, sessionId, common.DATA_FAMILY_NAME)
	assert.Equal(t, familyName, string(storedFamilyName))
}

func TestSaveYoB(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_yob_set, _ := fm.GetFlag("flag_yob_set")

	// Set the flag in the State
	mockState := state.NewState(108)
	mockState.SetFlag(flag_allow_update)

	expectedResult := resource.Result{}

	// Define test data
	yob := "1980"

	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(yob)); err != nil {
		t.Fatal(err)
	}

	expectedResult.FlagSet = []uint32{flag_yob_set}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
		flagManager:   fm.parser,
		st:            mockState,
	}

	// Call the method
	res, err := h.SaveYob(ctx, "save_yob", []byte(yob))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	// Verify that the DATA_YOB entry has been updated with the temporary value
	storedYob, _ := store.ReadEntry(ctx, sessionId, common.DATA_YOB)
	assert.Equal(t, yob, string(storedYob))
}

func TestSaveLocation(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_location_set, _ := fm.GetFlag("flag_location_set")

	// Set the flag in the State
	mockState := state.NewState(108)
	mockState.SetFlag(flag_allow_update)

	expectedResult := resource.Result{}

	// Define test data
	location := "Kilifi"

	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(location)); err != nil {
		t.Fatal(err)
	}

	expectedResult.FlagSet = []uint32{flag_location_set}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
		flagManager:   fm.parser,
		st:            mockState,
	}

	// Call the method
	res, err := h.SaveLocation(ctx, "save_location", []byte(location))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	// Verify that the DATA_LOCATION entry has been updated with the temporary value
	storedLocation, _ := store.ReadEntry(ctx, sessionId, common.DATA_LOCATION)
	assert.Equal(t, location, string(storedLocation))
}

func TestSaveOfferings(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_offerings_set, _ := fm.GetFlag("flag_offerings_set")

	// Set the flag in the State
	mockState := state.NewState(108)
	mockState.SetFlag(flag_allow_update)

	expectedResult := resource.Result{}

	// Define test data
	offerings := "Bananas"

	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(offerings)); err != nil {
		t.Fatal(err)
	}

	expectedResult.FlagSet = []uint32{flag_offerings_set}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
		flagManager:   fm.parser,
		st:            mockState,
	}

	// Call the method
	res, err := h.SaveOfferings(ctx, "save_offerings", []byte(offerings))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	// Verify that the DATA_OFFERINGS entry has been updated with the temporary value
	storedOfferings, _ := store.ReadEntry(ctx, sessionId, common.DATA_OFFERINGS)
	assert.Equal(t, offerings, string(storedOfferings))
}

func TestSaveGender(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	flag_gender_set, _ := fm.GetFlag("flag_gender_set")

	// Set the flag in the State
	mockState := state.NewState(108)
	mockState.SetFlag(flag_allow_update)

	// Define test cases
	tests := []struct {
		name            string
		input           []byte
		expectedGender  string
		executingSymbol string
	}{
		{
			name:            "Valid Male Input",
			input:           []byte("1"),
			expectedGender:  "male",
			executingSymbol: "set_male",
		},
		{
			name:            "Valid Female Input",
			input:           []byte("2"),
			expectedGender:  "female",
			executingSymbol: "set_female",
		},
		{
			name:            "Valid Unspecified Input",
			input:           []byte("3"),
			executingSymbol: "set_unspecified",
			expectedGender:  "unspecified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(tt.expectedGender)); err != nil {
				t.Fatal(err)
			}

			mockState.ExecPath = append(mockState.ExecPath, tt.executingSymbol)
			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: store,
				st:            mockState,
				flagManager:   fm.parser,
			}

			expectedResult := resource.Result{}

			// Call the method
			res, err := h.SaveGender(ctx, "save_gender", tt.input)

			expectedResult.FlagSet = []uint32{flag_gender_set}

			// Assert results
			assert.NoError(t, err)
			assert.Equal(t, expectedResult, res)

			// Verify that the DATA_GENDER entry has been updated with the temporary value
			storedGender, _ := store.ReadEntry(ctx, sessionId, common.DATA_GENDER)
			assert.Equal(t, tt.expectedGender, string(storedGender))
		})
	}
}

func TestSaveTemporaryPin(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		log.Fatal(err)
	}

	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

	// Create the Handlers instance with the mock flag manager
	h := &Handlers{
		flagManager:   fm.parser,
		userdataStore: store,
	}

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
			// Call the method
			res, err := h.SaveTemporaryPin(ctx, "save_pin", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should match expected result")
		})
	}
}

func TestCheckIdentifier(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name            string
		publicKey       []byte
		mockErr         error
		expectedContent string
		expectError     bool
	}{
		{
			name:            "Saved public Key",
			publicKey:       []byte("0xa8363"),
			mockErr:         nil,
			expectedContent: "0xa8363",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(tt.publicKey))
			if err != nil {
				t.Fatal(err)
			}

			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: store,
			}

			// Call the method
			res, err := h.CheckIdentifier(ctx, "check_identifier", nil)

			// Assert results
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, res.Content)
		})
	}
}

func TestGetSender(t *testing.T) {
	sessionId := "session123"
	ctx, _ := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	// Create the Handlers instance
	h := &Handlers{}

	// Call the method
	res, _ := h.GetSender(ctx, "get_sender", []byte(""))

	//Assert that the sessionId is what was set as the result content.
	assert.Equal(t, sessionId, res.Content)
}

func TestGetAmount(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	// Define test data
	amount := "0.03"
	activeSym := "SRF"

	err := store.WriteEntry(ctx, sessionId, common.DATA_AMOUNT, []byte(amount))
	if err != nil {
		t.Fatal(err)
	}

	err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_SYM, []byte(activeSym))
	if err != nil {
		t.Fatal(err)
	}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
	}

	// Call the method
	res, _ := h.GetAmount(ctx, "get_amount", []byte(""))

	formattedAmount := fmt.Sprintf("%s %s", amount, activeSym)

	//Assert that the retrieved amount is what was set as the content
	assert.Equal(t, formattedAmount, res.Content)
}

func TestGetRecipient(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	recepient := "0712345678"

	err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(recepient))
	if err != nil {
		t.Fatal(err)
	}

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: store,
	}

	// Call the method
	res, _ := h.GetRecipient(ctx, "get_recipient", []byte(""))

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
	if err != nil {
		log.Fatal(err)
	}

	flag_allow_update, _ := fm.parser.GetFlag("flag_allow_update")

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
	if err != nil {
		log.Fatal(err)
	}

	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

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
	if err != nil {
		log.Fatal(err)
	}

	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

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
	if err != nil {
		log.Fatal(err)
	}

	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")

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
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}

	// Create required mocks
	mockAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)
	flag_incorrect_pin, _ := fm.GetFlag("flag_incorrect_pin")
	flag_account_authorized, _ := fm.GetFlag("flag_account_authorized")
	flag_allow_update, _ := fm.GetFlag("flag_allow_update")

	// Set 1234 is the correct account pin
	accountPIN := "1234"

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
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
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACCOUNT_PIN, []byte(accountPIN))
			if err != nil {
				t.Fatal(err)
			}

			// Call the method under test
			res, err := h.Authorize(ctx, "authorize", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
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
	mockAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)
	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		accountService: mockAccountService,
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
		})
	}
}

func TestVerifyCreatePin(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}

	// Create required mocks
	mockAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	flag_pin_set, _ := fm.parser.GetFlag("flag_pin_set")

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte("1234"))
			if err != nil {
				t.Fatal(err)
			}

			// Call the method under test
			res, err := h.VerifyCreatePin(ctx, "verify_create_pin", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
		})
	}
}

func TestCheckAccountStatus(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	flag_account_success, _ := fm.GetFlag("flag_account_success")
	flag_account_pending, _ := fm.GetFlag("flag_account_pending")
	flag_api_error, _ := fm.GetFlag("flag_api_call_error")

	tests := []struct {
		name           string
		publicKey      []byte
		response       *models.TrackStatusResult
		expectedResult resource.Result
	}{
		{
			name:      "Test when account is on the Sarafu network",
			publicKey: []byte("TrackingId1234"),
			response: &models.TrackStatusResult{
				Active: true,
			},
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_account_success},
				FlagReset: []uint32{flag_api_error, flag_account_pending},
			},
		},
		{
			name:      "Test when the account is not yet on the sarafu network",
			publicKey: []byte("TrackingId1234"),
			response: &models.TrackStatusResult{
				Active: false,
			},
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_account_pending},
				FlagReset: []uint32{flag_api_error, flag_account_success},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccountService := new(mocks.MockAccountService)

			h := &Handlers{
				userdataStore:  store,
				accountService: mockAccountService,
				flagManager:    fm.parser,
			}

			err = store.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(tt.publicKey))
			if err != nil {
				t.Fatal(err)
			}

			mockAccountService.On("TrackAccountStatus", string(tt.publicKey)).Return(tt.response, nil)

			// Call the method under test
			res, _ := h.CheckAccountStatus(ctx, "check_account_status", []byte(""))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
		})
	}
}

func TestTransactionReset(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}

	flag_invalid_recipient, _ := fm.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := fm.GetFlag("flag_invalid_recipient_with_invite")

	mockAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
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
			// Call the method under test
			res, _ := h.TransactionReset(ctx, "transaction_reset", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
		})
	}
}

func TestResetTransactionAmount(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}

	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")

	mockAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name           string
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
			// Call the method under test
			res, _ := h.ResetTransactionAmount(ctx, "transaction_reset_amount", []byte(""))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
		})
	}
}

func TestInitiateTransaction(t *testing.T) {
	sessionId := "254712345678"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	account_authorized_flag, _ := fm.parser.GetFlag("flag_account_authorized")

	mockAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name             string
		TemporaryValue   []byte
		ActiveSym        []byte
		StoredAmount     []byte
		TransferAmount   string
		PublicKey        []byte
		Recipient        []byte
		ActiveDecimal    []byte
		ActiveAddress    []byte
		TransferResponse *models.TokenTransferResponse
		expectedResult   resource.Result
	}{
		{
			name:           "Test initiate transaction",
			TemporaryValue: []byte("0711223344"),
			ActiveSym:      []byte("SRF"),
			StoredAmount:   []byte("1.00"),
			TransferAmount: "1000000",
			PublicKey:      []byte("0X13242618721"),
			Recipient:      []byte("0x12415ass27192"),
			ActiveDecimal:  []byte("6"),
			ActiveAddress:  []byte("0xd4c288865Ce"),
			TransferResponse: &models.TokenTransferResponse{
				TrackingId: "1234567890",
			},
			expectedResult: resource.Result{
				FlagReset: []uint32{account_authorized_flag},
				Content:   "Your request has been sent. 0711223344 will receive 1.00 SRF from 254712345678.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(tt.TemporaryValue))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_SYM, []byte(tt.ActiveSym))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_AMOUNT, []byte(tt.StoredAmount))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(tt.PublicKey))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_RECIPIENT, []byte(tt.Recipient))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_DECIMAL, []byte(tt.ActiveDecimal))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_ADDRESS, []byte(tt.ActiveAddress))
			if err != nil {
				t.Fatal(err)
			}

			mockAccountService.On("TokenTransfer").Return(tt.TransferResponse, nil)

			// Call the method under test
			res, _ := h.InitiateTransaction(ctx, "transaction_reset_amount", []byte(""))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")
		})
	}
}

func TestQuit(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	mockAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		accountService: mockAccountService,
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

	sessionId := "session123"

	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")

	mockAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
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
			name:      "Test with invalid amount format",
			input:     []byte("0.02ms"),
			activeBal: []byte("5"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_amount},
				Content: "0.02ms",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_BAL, []byte(tt.activeBal))
			if err != nil {
				t.Fatal(err)
			}

			// Call the method under test
			res, _ := h.ValidateAmount(ctx, "test_validate_amount", tt.input)

			// Assert no errors occurred
			assert.NoError(t, err)

			// Assert the result matches the expected result
			assert.Equal(t, tt.expectedResult, res, "Expected result should match actual result")
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		log.Fatal(err)
	}

	sessionId := "session123"
	publicKey := "0X13242618721"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	flag_invalid_recipient, _ := fm.parser.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := fm.parser.GetFlag("flag_invalid_recipient_with_invite")

	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with invalid recepient",
			input: []byte("7?1234"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_recipient},
				Content: "7?1234",
			},
		},
		{
			name:  "Test with valid unregistered recepient",
			input: []byte("0712345678"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_recipient_with_invite},
				Content: "0712345678",
			},
		},
		{
			name:           "Test with valid registered recepient",
			input:          []byte("0711223344"),
			expectedResult: resource.Result{},
		},
		{
			name:           "Test with address",
			input:          []byte("0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9"),
			expectedResult: resource.Result{},
		},
		{
			name:           "Test with alias recepient",
			input:          []byte("alias123"),
			expectedResult: resource.Result{},
		},
	}

	// store a public key for the valid recipient
	err = store.WriteEntry(ctx, "+254711223344", common.DATA_PUBLIC_KEY, []byte(publicKey))
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccountService := new(mocks.MockAccountService)
			// Create the Handlers instance
			h := &Handlers{
				flagManager:    fm.parser,
				userdataStore:  store,
				accountService: mockAccountService,
			}

			aliasResponse := &dataserviceapi.AliasAddress{
				Address: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
			}

			mockAccountService.On("CheckAliasAddress", string(tt.input)).Return(aliasResponse, nil)

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
	ctx, store := InitializeTestStore(t)

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
			sessionId:      "session123",
			publicKey:      "0X98765432109",
			activeSym:      "ETH",
			activeBal:      "1.5",
			expectedResult: resource.Result{Content: "Balance: 1.50 ETH\n"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccountService := new(mocks.MockAccountService)
			ctx := context.WithValue(ctx, "SessionId", tt.sessionId)

			h := &Handlers{
				userdataStore:  store,
				accountService: mockAccountService,
			}

			err := store.WriteEntry(ctx, tt.sessionId, common.DATA_ACTIVE_SYM, []byte(tt.activeSym))
			if err != nil {
				t.Fatal(err)
			}
			err = store.WriteEntry(ctx, tt.sessionId, common.DATA_ACTIVE_BAL, []byte(tt.activeBal))
			if err != nil {
				t.Fatal(err)
			}

			res, err := h.CheckBalance(ctx, "check_balance", []byte(""))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res, "Result should match expected output")
			}

			mockAccountService.AssertExpectations(t)
		})
	}
}

func TestGetProfile(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)

	mockAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
		st:             mockState,
	}

	tests := []struct {
		name         string
		languageCode string
		keys         []common.DataTyp
		profileInfo  []string
		result       resource.Result
	}{
		{
			name:         "Test with full profile information in eng",
			keys:         []common.DataTyp{common.DATA_FAMILY_NAME, common.DATA_FIRST_NAME, common.DATA_GENDER, common.DATA_OFFERINGS, common.DATA_LOCATION, common.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "eng",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
					"John Doee", "Male", "49", "Kilifi", "Bananas",
				),
			},
		},
		{
			name:         "Test with with profile information in swa",
			keys:         []common.DataTyp{common.DATA_FAMILY_NAME, common.DATA_FIRST_NAME, common.DATA_GENDER, common.DATA_OFFERINGS, common.DATA_LOCATION, common.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "swa",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Jina: %s\nJinsia: %s\nUmri: %s\nEneo: %s\nUnauza: %s\n",
					"John Doee", "Male", "49", "Kilifi", "Bananas",
				),
			},
		},
		{
			name:         "Test with with profile information with language that is not yet supported",
			keys:         []common.DataTyp{common.DATA_FAMILY_NAME, common.DATA_FIRST_NAME, common.DATA_GENDER, common.DATA_OFFERINGS, common.DATA_LOCATION, common.DATA_YOB},
			profileInfo:  []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			languageCode: "nor",
			result: resource.Result{
				Content: fmt.Sprintf(
					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
					"John Doee", "Male", "49", "Kilifi", "Bananas",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = context.WithValue(ctx, "SessionId", sessionId)
			ctx = context.WithValue(ctx, "Language", lang.Language{
				Code: tt.languageCode,
			})
			for index, key := range tt.keys {
				err := store.WriteEntry(ctx, sessionId, key, []byte(tt.profileInfo[index]))
				if err != nil {
					t.Fatal(err)
				}
			}

			res, _ := h.GetProfileInfo(ctx, "get_profile_info", []byte(""))

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.result, "Result should contain profile information served back to user")
		})
	}
}

func TestVerifyNewPin(t *testing.T) {
	sessionId := "session123"

	fm, _ := NewFlagManager(flagsPath)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	mockAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		flagManager:    fm.parser,
		accountService: mockAccountService,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Call the function under test
			res, _ := h.VerifyNewPin(ctx, "verify_new_pin", tt.input)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")
		})
	}
}

func TestConfirmPin(t *testing.T) {
	sessionId := "session123"

	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, _ := NewFlagManager(flagsPath)
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	mockAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		userdataStore:  store,
		flagManager:    fm.parser,
		accountService: mockAccountService,
	}

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
			err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(tt.temporarypin))
			if err != nil {
				t.Fatal(err)
			}

			//Call the function under test
			res, _ := h.ConfirmPinChange(ctx, "confirm_pin_change", tt.temporarypin)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}
}

func TestFetchCommunityBalance(t *testing.T) {

	// Define test data
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)

	tests := []struct {
		name           string
		languageCode   string
		expectedResult resource.Result
	}{
		{
			name: "Test community balance content when language is english",
			expectedResult: resource.Result{
				Content: "Community Balance: 0.00",
			},
			languageCode: "eng",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAccountService := new(mocks.MockAccountService)
			mockState := state.NewState(16)

			h := &Handlers{
				userdataStore:  store,
				st:             mockState,
				accountService: mockAccountService,
			}
			ctx = context.WithValue(ctx, "SessionId", sessionId)
			ctx = context.WithValue(ctx, "Language", lang.Language{
				Code: tt.languageCode,
			})

			// Call the method
			res, _ := h.FetchCommunityBalance(ctx, "fetch_community_balance", []byte(""))

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should match expected result")
		})
	}
}

func TestSetDefaultVoucher(t *testing.T) {
	sessionId := "session123"
	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	flag_no_active_voucher, err := fm.GetFlag("flag_no_active_voucher")
	if err != nil {
		t.Logf(err.Error())
	}

	publicKey := "0X13242618721"

	tests := []struct {
		name           string
		vouchersResp   []dataserviceapi.TokenHoldings
		expectedResult resource.Result
	}{
		{
			name:         "Test no vouchers available",
			vouchersResp: []dataserviceapi.TokenHoldings{},
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_no_active_voucher},
			},
		},
		{
			name: "Test set default voucher when no active voucher is set",
			vouchersResp: []dataserviceapi.TokenHoldings{
				dataserviceapi.TokenHoldings{
					ContractAddress: "0x123",
					TokenSymbol:     "TOKEN1",
					TokenDecimals:   "18",
					Balance:         "100",
				},
			},
			expectedResult: resource.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccountService := new(mocks.MockAccountService)

			h := &Handlers{
				userdataStore:  store,
				accountService: mockAccountService,
				flagManager:    fm.parser,
			}

			err := store.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(publicKey))
			if err != nil {
				t.Fatal(err)
			}

			mockAccountService.On("FetchVouchers", string(publicKey)).Return(tt.vouchersResp, nil)

			res, err := h.SetDefaultVoucher(ctx, "set_default_voucher", []byte("some-input"))

			assert.NoError(t, err)

			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			mockAccountService.AssertExpectations(t)
		})
	}
}

func TestCheckVouchers(t *testing.T) {
	mockAccountService := new(mocks.MockAccountService)
	sessionId := "session123"
	publicKey := "0X13242618721"

	ctx, store := InitializeTestStore(t)
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	spdb := InitializeTestSubPrefixDb(t, ctx)

	h := &Handlers{
		userdataStore:  store,
		accountService: mockAccountService,
		prefixDb:       spdb,
	}

	err := store.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(publicKey))
	if err != nil {
		t.Fatal(err)
	}

	mockVouchersResponse := []dataserviceapi.TokenHoldings{
		{ContractAddress: "0xd4c288865Ce", TokenSymbol: "SRF", TokenDecimals: "6", Balance: "100"},
		{ContractAddress: "0x41c188d63Qa", TokenSymbol: "MILO", TokenDecimals: "4", Balance: "200"},
	}

	expectedSym := []byte("1:SRF\n2:MILO")

	mockAccountService.On("FetchVouchers", string(publicKey)).Return(mockVouchersResponse, nil)

	_, err = h.CheckVouchers(ctx, "check_vouchers", []byte(""))
	assert.NoError(t, err)

	// Read voucher sym data from the store
	voucherData, err := spdb.Get(ctx, common.ToBytes(common.DATA_VOUCHER_SYMBOLS))
	if err != nil {
		t.Fatal(err)
	}

	// assert that the data is stored correctly
	assert.Equal(t, expectedSym, voucherData)

	mockAccountService.AssertExpectations(t)
}

func TestGetVoucherList(t *testing.T) {
	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	spdb := InitializeTestSubPrefixDb(t, ctx)

	// Initialize Handlers
	h := &Handlers{
		prefixDb:             spdb,
		ReplaceSeparatorFunc: mockReplaceSeparator,
	}

	mockSyms := []byte("1:SRF\n2:MILO")

	// Put voucher sym data from the store
	err := spdb.Put(ctx, common.ToBytes(common.DATA_VOUCHER_SYMBOLS), mockSyms)
	if err != nil {
		t.Fatal(err)
	}

	expectedSyms := []byte("1: SRF\n2: MILO")

	res, err := h.GetVoucherList(ctx, "", []byte(""))

	assert.NoError(t, err)
	assert.Equal(t, res.Content, string(expectedSyms))
}

func TestViewVoucher(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	ctx, store := InitializeTestStore(t)
	sessionId := "session123"

	ctx = context.WithValue(ctx, "SessionId", sessionId)

	spdb := InitializeTestSubPrefixDb(t, ctx)

	h := &Handlers{
		userdataStore: store,
		flagManager:   fm.parser,
		prefixDb:      spdb,
	}

	// Define mock voucher data
	mockData := map[common.DataTyp][]byte{
		common.DATA_VOUCHER_SYMBOLS:   []byte("1:SRF\n2:MILO"),
		common.DATA_VOUCHER_BALANCES:  []byte("1:100\n2:200"),
		common.DATA_VOUCHER_DECIMALS:  []byte("1:6\n2:4"),
		common.DATA_VOUCHER_ADDRESSES: []byte("1:0xd4c288865Ce\n2:0x41c188d63Qa"),
	}

	// Put the data
	for key, value := range mockData {
		err = spdb.Put(ctx, []byte(common.ToBytes(key)), []byte(value))
		if err != nil {
			t.Fatal(err)
		}
	}

	res, err := h.ViewVoucher(ctx, "view_voucher", []byte("1"))
	assert.NoError(t, err)
	assert.Equal(t, res.Content, "Symbol: SRF\nBalance: 100")
}

func TestSetVoucher(t *testing.T) {
	ctx, store := InitializeTestStore(t)
	sessionId := "session123"

	ctx = context.WithValue(ctx, "SessionId", sessionId)

	h := &Handlers{
		userdataStore: store,
	}

	// Define the temporary voucher data
	tempData := &dataserviceapi.TokenHoldings{
		TokenSymbol:     "SRF",
		Balance:         "200",
		TokenDecimals:   "6",
		ContractAddress: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	expectedData := fmt.Sprintf("%s,%s,%s,%s", tempData.TokenSymbol, tempData.Balance, tempData.TokenDecimals, tempData.ContractAddress)

	// store the expectedData
	if err := store.WriteEntry(ctx, sessionId, common.DATA_TEMPORARY_VALUE, []byte(expectedData)); err != nil {
		t.Fatal(err)
	}

	res, err := h.SetVoucher(ctx, "set_voucher", []byte(""))

	assert.NoError(t, err)

	assert.Equal(t, string(tempData.TokenSymbol), res.Content)
}

func TestGetVoucherDetails(t *testing.T) {
	ctx, store := InitializeTestStore(t)
	fm, err := NewFlagManager(flagsPath)
	if err != nil {
		t.Logf(err.Error())
	}
	mockAccountService := new(mocks.MockAccountService)

	sessionId := "session123"
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	expectedResult := resource.Result{}

	tokA_AAddress := "0x0000000000000000000000000000000000000000"

	h := &Handlers{
		userdataStore:  store,
		flagManager:    fm.parser,
		accountService: mockAccountService,
	}
	err = store.WriteEntry(ctx, sessionId, common.DATA_ACTIVE_ADDRESS, []byte(tokA_AAddress))
	if err != nil {
		t.Fatal(err)
	}
	tokenDetails := &models.VoucherDataResult{
		TokenName:      "Token A",
		TokenSymbol:    "TOKA",
		TokenLocation:  "Kilifi,Kenya",
		TokenCommodity: "Farming",
	}
	expectedResult.Content = fmt.Sprintf(
		"Name: %s\nSymbol: %s\nCommodity: %s\nLocation: %s", tokenDetails.TokenName, tokenDetails.TokenSymbol, tokenDetails.TokenCommodity, tokenDetails.TokenLocation,
	)
	mockAccountService.On("VoucherData", string(tokA_AAddress)).Return(tokenDetails, nil)

	res, err := h.GetVoucherDetails(ctx, "SessionId", []byte(""))
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)
}
