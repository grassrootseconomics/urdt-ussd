package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	httpserver "git.grassecon.net/urdt/ussd/internal/http"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

type atRequestParser struct{}

func (arp *atRequestParser) GetSessionId(rq any) (string, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return "", handlers.ErrInvalidRequest
	}
	if err := rqv.ParseForm(); err != nil {
		return "", fmt.Errorf("failed to parse form data: %v", err)
	}

	phoneNumber := rqv.FormValue("phoneNumber")
	if phoneNumber == "" {
		return "", fmt.Errorf("no phone number found")
	}

	return phoneNumber, nil
}

func (arp *atRequestParser) GetInput(rq any) ([]byte, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return nil, handlers.ErrInvalidRequest
	}
	if err := rqv.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form data: %v", err)
	}

	text := rqv.FormValue("text")

	parts := strings.Split(text, "*")
	if len(parts) == 0 {
		return nil, fmt.Errorf("no input found")
	}

	return []byte(parts[len(parts)-1]), nil
}

func main() {
	var dbDir string
	var resourceDir string
	var size uint
	var engineDebug bool
	var host string
	var port uint
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", "127.0.0.1", "http host")
	flag.UintVar(&port, "p", 7123, "http port")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size)

	ctx := context.Background()
	pfp := path.Join(scriptDir, "pp.csv")

	cfg := engine.Config{
		Root:       "root",
		OutputSize: uint32(size),
		FlagCount:  uint32(16),
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

	lhs, err := handlers.NewLocalHandlerService(pfp, true, dbResource, cfg, rs)
	lhs.SetDataStore(&userdataStore)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	accountService := server.AccountService{}
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

	rp := &atRequestParser{}
	bsh := handlers.NewBaseSessionHandler(cfg, rs, stateStore, userdataStore, rp, hl)
	sh := httpserver.NewATSessionHandler(bsh)
	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, strconv.Itoa(int(port))),
		Handler: sh,
	}
	s.RegisterOnShutdown(sh.Shutdown)

	cint := make(chan os.Signal)
	cterm := make(chan os.Signal)
	signal.Notify(cint, os.Interrupt, syscall.SIGINT)
	signal.Notify(cterm, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case _ = <-cint:
		case _ = <-cterm:
		}
		s.Shutdown(ctx)
	}()
	err = s.ListenAndServe()
	if err != nil {
		logg.Infof("Server closed with error", "err", err)
	}
}
