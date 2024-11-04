package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/remote"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

func init() {
	initializers.LoadEnvVariables()
}

type asyncRequestParser struct {
	sessionId string
	input     []byte
}

func (p *asyncRequestParser) GetSessionId(r any) (string, error) {
	return p.sessionId, nil
}

func (p *asyncRequestParser) GetInput(r any) ([]byte, error) {
	return p.input, nil
}

func main() {
	config.LoadConfig()

	var sessionId string
	var dbDir string
	var resourceDir string
	var size uint
	var database string
	var engineDebug bool
	var host string
	var port uint
	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.StringVar(&database, "db", "gdbm", "database to be used")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", initializers.GetEnv("HOST", "127.0.0.1"), "http host")
	flag.UintVar(&port, "p", initializers.GetEnvUint("PORT", 7123), "http port")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size, "sessionId", sessionId)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "Database", database)
	pfp := path.Join(scriptDir, "pp.csv")

	cfg := engine.Config{
		Root:       "root",
		OutputSize: uint32(size),
		FlagCount:  uint32(128),
	}

	if engineDebug {
		cfg.EngineDebug = true
	}

	menuStorageService := storage.NewMenuStorageService(dbDir, resourceDir)
	rs, err := menuStorageService.GetResource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = menuStorageService.EnsureDbDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userdataStore, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer userdataStore.Close()

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(ctx, pfp, true, dbResource, cfg, rs)
	lhs.SetDataStore(&userdataStore)
	accountService := remote.AccountService{}

	hl, err := lhs.GetHandler(&accountService)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	stateStore, err := menuStorageService.GetStateStore(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer stateStore.Close()

	rp := &asyncRequestParser{
		sessionId: sessionId,
	}
	sh := handlers.NewBaseSessionHandler(cfg, rs, stateStore, userdataStore, rp, hl)
	cfg.SessionId = sessionId
	rqs := handlers.RequestSession{
		Ctx:    ctx,
		Writer: os.Stdout,
		Config: cfg,
	}

	cint := make(chan os.Signal)
	cterm := make(chan os.Signal)
	signal.Notify(cint, os.Interrupt, syscall.SIGINT)
	signal.Notify(cterm, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case _ = <-cint:
		case _ = <-cterm:
		}
		sh.Shutdown()
	}()

	for true {
		rqs, err = sh.Process(rqs)
		if err != nil {
			logg.ErrorCtxf(ctx, "error in process: %v", "err", err)
			fmt.Errorf("error in process: %v", err)
			os.Exit(1)
		}
		rqs, err = sh.Output(rqs)
		if err != nil {
			logg.ErrorCtxf(ctx, "error in output: %v", "err", err)
			fmt.Errorf("error in output: %v", err)
			os.Exit(1)
		}
		rqs, err = sh.Reset(rqs)
		if err != nil {
			logg.ErrorCtxf(ctx, "error in reset: %v", "err", err)
			fmt.Errorf("error in reset: %v", err)
			os.Exit(1)
		}
		fmt.Println("")
		_, err = fmt.Scanln(&rqs.Input)
		if err != nil {
			logg.ErrorCtxf(ctx, "error in input", "err", err)
			fmt.Errorf("error in input: %v", err)
			os.Exit(1)
		}
	}
}
