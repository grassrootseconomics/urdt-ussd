package handlers

import (
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

type BaseSessionHandler struct {
	cfgTemplate engine.Config
	rp RequestParser
	rs resource.Resource
	hn *ussd.Handlers
	provider storage.StorageProvider
}

func NewBaseSessionHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp RequestParser, hn *ussd.Handlers) *BaseSessionHandler {
	return &BaseSessionHandler{
		cfgTemplate: cfg,
		rs: rs,
		hn: hn,
		rp: rp,
		provider: storage.NewSimpleStorageProvider(stateDb, userdataDb),
	}
}

func(f* BaseSessionHandler) Shutdown() {
	err := f.provider.Close()
	if err != nil {
		logg.Errorf("handler shutdown error", "err", err)
	}
}

func(f *BaseSessionHandler) GetEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) engine.Engine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func(f *BaseSessionHandler) Process(rqs RequestSession) (RequestSession, error) {
	var r bool
	var err error
	var ok bool
	
	logg.InfoCtxf(rqs.Ctx, "new request",  rqs)

	rqs.Storage, err = f.provider.Get(rqs.Config.SessionId)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "storage error", "err", err)
		return rqs, ErrStorage
	}

	f.hn = f.hn.WithPersister(rqs.Storage.Persister)
	eni := f.GetEngine(rqs.Config, f.rs, rqs.Storage.Persister)
	en, ok := eni.(*engine.DefaultEngine)
	if !ok {
		return rqs, ErrEngineType
	}
	en = en.WithFirst(f.hn.Init)
	if rqs.Config.EngineDebug {
		en = en.WithDebug(nil)
	}
	rqs.Engine = en

	r, err = rqs.Engine.Init(rqs.Ctx)
	if err != nil {
		return rqs, err
	}

	if r && len(rqs.Input) > 0 {
		r, err = rqs.Engine.Exec(rqs.Ctx, rqs.Input)
	}
	if err != nil {
		return rqs, err
	}

	rqs.Continue = r 
	return rqs, nil
}

func(f *BaseSessionHandler) Output(rqs RequestSession) (RequestSession,  error) {
	var err error
	_, err = rqs.Engine.WriteResult(rqs.Ctx, rqs.Writer)
	return rqs, err
}

func(f *BaseSessionHandler) Reset(rqs RequestSession) (RequestSession, error) {
	defer f.provider.Put(rqs.Config.SessionId, rqs.Storage)
	return rqs, rqs.Engine.Finish()
}

func(f *BaseSessionHandler) GetConfig() engine.Config {
	return f.cfgTemplate
}

func(f *BaseSessionHandler) GetRequestParser() RequestParser {
	return f.rp
}
