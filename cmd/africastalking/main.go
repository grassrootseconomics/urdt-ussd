package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
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

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	httpserver "git.grassecon.net/urdt/ussd/internal/http"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/remote"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")

	build = "dev"
)

func init() {
	initializers.LoadEnvVariables()
}

type atRequestParser struct{}

func (arp *atRequestParser) GetSessionId(rq any) (string, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		log.Println("got an invalid request:", rq)
		return "", handlers.ErrInvalidRequest
	}

	// Capture body (if any) for logging
	body, err := io.ReadAll(rqv.Body)
	if err != nil {
		log.Println("failed to read request body:", err)
		return "", fmt.Errorf("failed to read request body: %v", err)
	}
	// Reset the body for further reading
	rqv.Body = io.NopCloser(bytes.NewReader(body))

	// Log the body as JSON
	bodyLog := map[string]string{"body": string(body)}
	logBytes, err := json.Marshal(bodyLog)
	if err != nil {
		log.Println("failed to marshal request body:", err)
	} else {
		log.Println("Received request:", string(logBytes))
	}

	if err := rqv.ParseForm(); err != nil {
		log.Println("failed to parse form data: %v", err)
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
	config.LoadConfig()

	var dbDir string
	var resourceDir string
	var size uint
	var database string
	var engineDebug bool
	var host string
	var port uint
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.StringVar(&database, "db", "gdbm", "database to be used")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", initializers.GetEnv("HOST", "127.0.0.1"), "http host")
	flag.UintVar(&port, "p", initializers.GetEnvUint("PORT", 7123), "http port")
	flag.Parse()

	logg.Infof("start command", "build", build, "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size)

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
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	lhs.SetDataStore(&userdataStore)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

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

	rp := &atRequestParser{}
	bsh := handlers.NewBaseSessionHandler(cfg, rs, stateStore, userdataStore, rp, hl)
	sh := httpserver.NewATSessionHandler(bsh)

	mux := http.NewServeMux()
	mux.Handle(initializers.GetEnv("AT_ENDPOINT", "/"), sh)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, strconv.Itoa(int(port))),
		Handler: mux,
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
