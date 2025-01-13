package httpmocks

import "net/http"

// MockWriter implements a mock io.Writer for testing
type MockWriter struct {
	WriteStringCalled bool
	WrittenString     string
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockWriter) WriteString(s string) (n int, err error) {
	m.WriteStringCalled = true
	m.WrittenString = s
	return len(s), nil
}

func (m *MockWriter) Header() http.Header {
	return http.Header{}
}

func (m *MockWriter) WriteHeader(statusCode int) {}