package http

import (
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/logging"

	"git.grassecon.net/urdt/ussd/internal/handlers"
)

var (
	logg = logging.NewVanilla().WithDomain("httpserver")
)

type SessionHandler struct {
	handlers.RequestHandler
}

func ToSessionHandler(h handlers.RequestHandler) *SessionHandler {
	return &SessionHandler{
		RequestHandler: h,
	}
}

func (f *SessionHandler) WriteError(w http.ResponseWriter, code int, err error) {
	s := err.Error()
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	_, err = w.Write([]byte(s))
	if err != nil {
		logg.Errorf("error writing error!!", "err", err, "olderr", s)
		w.WriteHeader(500)
	}
}

func (f *SessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error
	var perr error

	rqs := handlers.RequestSession{
		Ctx:    req.Context(),
		Writer: w,
	}

	rp := f.GetRequestParser()
	cfg := f.GetConfig()
	cfg.SessionId, err = rp.GetSessionId(req.Context(), req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		f.WriteError(w, 400, err)
	}
	rqs.Config = cfg
	rqs.Input, err = rp.GetInput(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		f.WriteError(w, 400, err)
		return
	}

	rqs, err = f.Process(rqs)
	switch err {
	case handlers.ErrStorage:
		code = 500
	case handlers.ErrEngineInit:
		code = 500
	case handlers.ErrEngineExec:
		code = 500
	default:
		code = 200
	}

	if code != 200 {
		f.WriteError(w, 500, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	rqs, err = f.Output(rqs)
	rqs, perr = f.Reset(rqs)
	if err != nil {
		f.WriteError(w, 500, err)
		return
	}
	if perr != nil {
		f.WriteError(w, 500, perr)
		return
	}
}
