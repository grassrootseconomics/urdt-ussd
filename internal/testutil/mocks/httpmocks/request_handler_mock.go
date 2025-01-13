package httpmocks

import (
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers"
)

// MockRequestHandler implements handlers.RequestHandler interface for testing
type MockRequestHandler struct {
	ProcessFunc          func(handlers.RequestSession) (handlers.RequestSession, error)
	GetConfigFunc        func() engine.Config
	GetEngineFunc        func(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine
	OutputFunc           func(rs handlers.RequestSession) (handlers.RequestSession, error)
	ResetFunc            func(rs handlers.RequestSession) (handlers.RequestSession, error)
	ShutdownFunc         func()
	GetRequestParserFunc func() handlers.RequestParser
}

func (m *MockRequestHandler) Process(rqs handlers.RequestSession) (handlers.RequestSession, error) {
	return m.ProcessFunc(rqs)
}

func (m *MockRequestHandler) GetConfig() engine.Config {
	return m.GetConfigFunc()
}

func (m *MockRequestHandler) GetEngine(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine {
	return m.GetEngineFunc(cfg, rs, pe)
}

func (m *MockRequestHandler) Output(rs handlers.RequestSession) (handlers.RequestSession, error) {
	return m.OutputFunc(rs)
}

func (m *MockRequestHandler) Reset(rs handlers.RequestSession) (handlers.RequestSession, error) {
	return m.ResetFunc(rs)
}

func (m *MockRequestHandler) Shutdown() {
	m.ShutdownFunc()
}

func (m *MockRequestHandler) GetRequestParser() handlers.RequestParser {
	return m.GetRequestParserFunc()
}
