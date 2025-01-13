package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.defalsify.org/vise.git/engine"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/testutil/mocks/httpmocks"
)

// invalidRequestType is a custom type to test invalid request scenarios
type invalidRequestType struct{}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestSessionHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		input          []byte
		parserErr      error
		processErr     error
		outputErr      error
		resetErr       error
		expectedStatus int
	}{
		{
			name:           "Success",
			sessionID:      "123",
			input:          []byte("test input"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Session ID",
			sessionID:      "",
			parserErr:      handlers.ErrSessionMissing,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Process Error",
			sessionID:      "123",
			input:          []byte("test input"),
			processErr:     handlers.ErrStorage,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Output Error",
			sessionID:      "123",
			input:          []byte("test input"),
			outputErr:      errors.New("output error"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Reset Error",
			sessionID:      "123",
			input:          []byte("test input"),
			resetErr:       errors.New("reset error"),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequestParser := &httpmocks.MockRequestParser{
				GetSessionIdFunc: func(any) (string, error) {
					return tt.sessionID, tt.parserErr
				},
				GetInputFunc: func(any) ([]byte, error) {
					return tt.input, nil
				},
			}

			mockRequestHandler := &httpmocks.MockRequestHandler{
				ProcessFunc: func(rs handlers.RequestSession) (handlers.RequestSession, error) {
					return rs, tt.processErr
				},
				OutputFunc: func(rs handlers.RequestSession) (handlers.RequestSession, error) {
					return rs, tt.outputErr
				},
				ResetFunc: func(rs handlers.RequestSession) (handlers.RequestSession, error) {
					return rs, tt.resetErr
				},
				GetRequestParserFunc: func() handlers.RequestParser {
					return mockRequestParser
				},
				GetConfigFunc: func() engine.Config {
					return engine.Config{}
				},
			}

			sessionHandler := ToSessionHandler(mockRequestHandler)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.input))
			req.Header.Set("X-Vise-Session", tt.sessionID)

			rr := httptest.NewRecorder()

			sessionHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func TestSessionHandler_WriteError(t *testing.T) {
	handler := &SessionHandler{}
	mockWriter := &httpmocks.MockWriter{}
	err := errors.New("test error")

	handler.WriteError(mockWriter, http.StatusBadRequest, err)

	if mockWriter.WrittenString != "" {
		t.Errorf("Expected empty body, got %s", mockWriter.WrittenString)
	}
}

func TestDefaultRequestParser_GetSessionId(t *testing.T) {
	tests := []struct {
		name          string
		request       any
		expectedID    string
		expectedError error
	}{
		{
			name: "Valid Session ID",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("X-Vise-Session", "123456")
				return req
			}(),
			expectedID:    "123456",
			expectedError: nil,
		},
		{
			name:          "Missing Session ID",
			request:       httptest.NewRequest(http.MethodPost, "/", nil),
			expectedID:    "",
			expectedError: handlers.ErrSessionMissing,
		},
		{
			name:          "Invalid Request Type",
			request:       invalidRequestType{},
			expectedID:    "",
			expectedError: handlers.ErrInvalidRequest,
		},
	}

	parser := &DefaultRequestParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parser.GetSessionId(context.Background(),tt.request)

			if id != tt.expectedID {
				t.Errorf("Expected session ID %s, got %s", tt.expectedID, id)
			}

			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestDefaultRequestParser_GetInput(t *testing.T) {
	tests := []struct {
		name          string
		request       any
		expectedInput []byte
		expectedError error
	}{
		{
			name: "Valid Input",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("test input"))
			}(),
			expectedInput: []byte("test input"),
			expectedError: nil,
		},
		{
			name:          "Empty Input",
			request:       httptest.NewRequest(http.MethodPost, "/", nil),
			expectedInput: []byte{},
			expectedError: nil,
		},
		{
			name:          "Invalid Request Type",
			request:       invalidRequestType{},
			expectedInput: nil,
			expectedError: handlers.ErrInvalidRequest,
		},
		{
			name: "Read Error",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", &errorReader{})
			}(),
			expectedInput: nil,
			expectedError: errors.New("read error"),
		},
	}

	parser := &DefaultRequestParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := parser.GetInput(tt.request)

			if !bytes.Equal(input, tt.expectedInput) {
				t.Errorf("Expected input %s, got %s", tt.expectedInput, input)
			}

			if err != tt.expectedError && (err == nil || err.Error() != tt.expectedError.Error()) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
