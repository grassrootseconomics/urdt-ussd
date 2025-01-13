package handlers

import (
	"context"
	"errors"
	"io"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg = logging.NewVanilla().WithDomain("handlers")
)

var (
	ErrInvalidRequest = errors.New("invalid request for context")
	ErrSessionMissing = errors.New("missing session")
	ErrInvalidInput   = errors.New("invalid input")
	ErrStorage        = errors.New("storage retrieval fail")
	ErrEngineType     = errors.New("incompatible engine")
	ErrEngineInit     = errors.New("engine init fail")
	ErrEngineExec     = errors.New("engine exec fail")
)

type RequestSession struct {
	Ctx      context.Context
	Config   engine.Config
	Engine   engine.Engine
	Input    []byte
	Storage  *storage.Storage
	Writer   io.Writer
	Continue bool
}

// TODO: seems like can remove this.
type RequestParser interface {
	GetSessionId(context context.Context, rq any) (string, error)
	GetInput(rq any) ([]byte, error)
}

type RequestHandler interface {
	GetConfig() engine.Config
	GetRequestParser() RequestParser
	GetEngine(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine
	Process(rs RequestSession) (RequestSession, error)
	Output(rs RequestSession) (RequestSession, error)
	Reset(rs RequestSession) (RequestSession, error)
	Shutdown()
}
