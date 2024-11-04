package httpmocks

import (
	"context"
	"io"
)

// MockEngine implements the engine.Engine interface for testing
type MockEngine struct {
	InitFunc   func(context.Context) (bool, error)
	ExecFunc   func(context.Context, []byte) (bool, error)
	FlushFunc  func(context.Context, io.Writer) (int, error)
	FinishFunc func() error
}

func (m *MockEngine) Init(ctx context.Context) (bool, error) {
	return m.InitFunc(ctx)
}

func (m *MockEngine) Exec(ctx context.Context, input []byte) (bool, error) {
	return m.ExecFunc(ctx, input)
}

func (m *MockEngine) Flush(ctx context.Context, w io.Writer) (int, error) {
	return m.FlushFunc(ctx, w)
}

func (m *MockEngine) Finish() error {
	return m.FinishFunc()
}
