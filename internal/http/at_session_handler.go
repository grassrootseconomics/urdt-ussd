package http

import (
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/internal/handlers"
)

type ATSessionHandler struct {
	*SessionHandler
}

func NewATSessionHandler(h handlers.RequestHandler) *ATSessionHandler {
	return &ATSessionHandler{
		SessionHandler: ToSessionHandler(h),
	}
}

func (ash *ATSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error

	rqs := handlers.RequestSession{
		Ctx:    req.Context(),
		Writer: w,
	}

	rp := ash.GetRequestParser()
	cfg := ash.GetConfig()
	cfg.SessionId, err = rp.GetSessionId(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		ash.writeError(w, 400, err)
		return
	}
	rqs.Config = cfg
	rqs.Input, err = rp.GetInput(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		ash.writeError(w, 400, err)
		return
	}

	rqs, err = ash.Process(rqs) 
	switch err {
	case nil: // set code to 200 if no err
		code = 200
	case handlers.ErrStorage, handlers.ErrEngineInit, handlers.ErrEngineExec, handlers.ErrEngineType:
		code = 500
	default:
		code = 500
	}

	if code != 200 {
		ash.writeError(w, 500, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	rqs, err = ash.Output(rqs)
	if err != nil {
		ash.writeError(w, 500, err)
		return
	}

	rqs, err = ash.Reset(rqs)
	if err != nil {
		ash.writeError(w, 500, err)
		return
	}
}

func (ash *ATSessionHandler) Output(rqs handlers.RequestSession) (handlers.RequestSession, error) {
	var err error
	var prefix string

	if rqs.Continue {
		prefix = "CON "
	} else {
		prefix = "END "
	}

	_, err = io.WriteString(rqs.Writer, prefix)
	if err != nil {
		return rqs, err
	}

	_, err = rqs.Engine.WriteResult(rqs.Ctx, rqs.Writer)
	return rqs, err
}