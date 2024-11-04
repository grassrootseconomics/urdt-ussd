package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestNewATSessionHandler(t *testing.T) {
	mockHandler := &httpmocks.MockRequestHandler{}
	ash := NewATSessionHandler(mockHandler)

	if ash == nil {
		t.Fatal("NewATSessionHandler returned nil")
	}

	if ash.SessionHandler == nil {
		t.Fatal("SessionHandler is nil")
	}
}

func TestATSessionHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*httpmocks.MockRequestHandler, *httpmocks.MockRequestParser, *httpmocks.MockEngine)
		formData       url.Values
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful request",
			setupMocks: func(mh *httpmocks.MockRequestHandler, mrp *httpmocks.MockRequestParser, me *httpmocks.MockEngine) {
				mrp.GetSessionIdFunc = func(rq any) (string, error) {
					req := rq.(*http.Request)
					return req.FormValue("phoneNumber"), nil
				}
				mrp.GetInputFunc = func(rq any) ([]byte, error) {
					req := rq.(*http.Request)
					text := req.FormValue("text")
					parts := strings.Split(text, "*")
					return []byte(parts[len(parts)-1]), nil
				}
				mh.ProcessFunc = func(rqs handlers.RequestSession) (handlers.RequestSession, error) {
					rqs.Continue = true
					rqs.Engine = me
					return rqs, nil
				}
				mh.GetConfigFunc = func() engine.Config { return engine.Config{} }
				mh.GetRequestParserFunc = func() handlers.RequestParser { return mrp }
				mh.OutputFunc = func(rs handlers.RequestSession) (handlers.RequestSession, error) { return rs, nil }
				mh.ResetFunc = func(rs handlers.RequestSession) (handlers.RequestSession, error) { return rs, nil }
				me.FlushFunc = func(context.Context, io.Writer) (int, error) { return 0, nil }
			},
			formData: url.Values{
				"phoneNumber": []string{"+1234567890"},
				"text":        []string{"1*2*3"},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "CON ",
		},
		{
			name: "GetSessionId error",
			setupMocks: func(mh *httpmocks.MockRequestHandler, mrp *httpmocks.MockRequestParser, me *httpmocks.MockEngine) {
				mrp.GetSessionIdFunc = func(rq any) (string, error) {
					return "", errors.New("no phone number found")
				}
				mh.GetConfigFunc = func() engine.Config { return engine.Config{} }
				mh.GetRequestParserFunc = func() handlers.RequestParser { return mrp }
			},
			formData: url.Values{
				"text": []string{"1*2*3"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "GetInput error",
			setupMocks: func(mh *httpmocks.MockRequestHandler, mrp *httpmocks.MockRequestParser, me *httpmocks.MockEngine) {
				mrp.GetSessionIdFunc = func(rq any) (string, error) {
					req := rq.(*http.Request)
					return req.FormValue("phoneNumber"), nil
				}
				mrp.GetInputFunc = func(rq any) ([]byte, error) {
					return nil, errors.New("no input found")
				}
				mh.GetConfigFunc = func() engine.Config { return engine.Config{} }
				mh.GetRequestParserFunc = func() handlers.RequestParser { return mrp }
			},
			formData: url.Values{
				"phoneNumber": []string{"+1234567890"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "Process error",
			setupMocks: func(mh *httpmocks.MockRequestHandler, mrp *httpmocks.MockRequestParser, me *httpmocks.MockEngine) {
				mrp.GetSessionIdFunc = func(rq any) (string, error) {
					req := rq.(*http.Request)
					return req.FormValue("phoneNumber"), nil
				}
				mrp.GetInputFunc = func(rq any) ([]byte, error) {
					req := rq.(*http.Request)
					text := req.FormValue("text")
					parts := strings.Split(text, "*")
					return []byte(parts[len(parts)-1]), nil
				}
				mh.ProcessFunc = func(rqs handlers.RequestSession) (handlers.RequestSession, error) {
					return rqs, handlers.ErrStorage
				}
				mh.GetConfigFunc = func() engine.Config { return engine.Config{} }
				mh.GetRequestParserFunc = func() handlers.RequestParser { return mrp }
			},
			formData: url.Values{
				"phoneNumber": []string{"+1234567890"},
				"text":        []string{"1*2*3"},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := &httpmocks.MockRequestHandler{}
			mockRequestParser := &httpmocks.MockRequestParser{}
			mockEngine := &httpmocks.MockEngine{}
			tt.setupMocks(mockHandler, mockRequestParser, mockEngine)

			ash := NewATSessionHandler(mockHandler)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			ash.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestATSessionHandler_Output(t *testing.T) {
	tests := []struct {
		name           string
		input          handlers.RequestSession
		expectedPrefix string
		expectedError  bool
	}{
		{
			name: "Continue true",
			input: handlers.RequestSession{
				Continue: true,
				Engine: &httpmocks.MockEngine{
					FlushFunc: func(context.Context, io.Writer) (int, error) {
						return 0, nil
					},
				},
				Writer: &httpmocks.MockWriter{},
			},
			expectedPrefix: "CON ",
			expectedError:  false,
		},
		{
			name: "Continue false",
			input: handlers.RequestSession{
				Continue: false,
				Engine: &httpmocks.MockEngine{
					FlushFunc: func(context.Context, io.Writer) (int, error) {
						return 0, nil
					},
				},
				Writer: &httpmocks.MockWriter{},
			},
			expectedPrefix: "END ",
			expectedError:  false,
		},
		{
			name: "Flush error",
			input: handlers.RequestSession{
				Continue: true,
				Engine: &httpmocks.MockEngine{
					FlushFunc: func(context.Context, io.Writer) (int, error) {
						return 0, errors.New("write error")
					},
				},
				Writer: &httpmocks.MockWriter{},
			},
			expectedPrefix: "CON ",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ash := &ATSessionHandler{}
			_, err := ash.Output(tt.input)

			if tt.expectedError && err == nil {
				t.Error("Expected an error, but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			mw := tt.input.Writer.(*httpmocks.MockWriter)
			if !mw.WriteStringCalled {
				t.Error("WriteString was not called")
			}

			if mw.WrittenString != tt.expectedPrefix {
				t.Errorf("Expected prefix %q, got %q", tt.expectedPrefix, mw.WrittenString)
			}
		})
	}
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

func TestSessionHandler_writeError(t *testing.T) {
	handler := &SessionHandler{}
	mockWriter := &httpmocks.MockWriter{}
	err := errors.New("test error")

	handler.writeError(mockWriter, http.StatusBadRequest, err)

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
			id, err := parser.GetSessionId(tt.request)

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
