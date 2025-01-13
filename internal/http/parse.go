package http

import (
	"context"
	"io/ioutil"
	"net/http"

	"git.grassecon.net/urdt/ussd/internal/handlers"
)

type DefaultRequestParser struct {
}

func (rp *DefaultRequestParser) GetSessionId(ctx context.Context, rq any) (string, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return "", handlers.ErrInvalidRequest
	}
	v := rqv.Header.Get("X-Vise-Session")
	if v == "" {
		return "", handlers.ErrSessionMissing
	}
	return v, nil
}

func (rp *DefaultRequestParser) GetInput(rq any) ([]byte, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return nil, handlers.ErrInvalidRequest
	}
	defer rqv.Body.Close()
	v, err := ioutil.ReadAll(rqv.Body)
	if err != nil {
		return nil, err
	}
	return v, nil
}
