package httpmocks

// MockRequestParser implements the handlers.RequestParser interface for testing
type MockRequestParser struct {
	GetSessionIdFunc func(any) (string, error)
	GetInputFunc     func(any) ([]byte, error)
}

func (m *MockRequestParser) GetSessionId(rq any) (string, error) {
	return m.GetSessionIdFunc(rq)
}

func (m *MockRequestParser) GetInput(rq any) ([]byte, error) {
	return m.GetInputFunc(rq)
}
