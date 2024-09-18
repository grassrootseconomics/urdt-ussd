package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"

	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
	"git.grassecon.net/urdt/ussd/internal/handlers"
)

var (
	logg = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

type asyncRequestParser struct {
	sessionId string
	input []byte
}

func(p *asyncRequestParser) GetSessionId(r any) (string, error) {
	return p.sessionId, nil
}

func(p *asyncRequestParser) GetInput(r any) ([]byte, error) {
	return p.input, nil
}

func getFlags(fp string, debug bool) (*asm.FlagParser, error) {
	flagParser := asm.NewFlagParser().WithDebug()
	_, err := flagParser.Load(fp)
	if err != nil {
		return nil, err
	}
	return flagParser, nil
}

func getHandler(appFlags *asm.FlagParser, rs *resource.DbResource, userdataStore db.Db) (*ussd.Handlers, error) {

	ussdHandlers, err := ussd.NewHandlers(appFlags, userdataStore)
	if err != nil {
		return nil, err
	}
	rs.AddLocalFunc("set_language", ussdHandlers.SetLanguage)
	rs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	rs.AddLocalFunc("save_pin", ussdHandlers.SavePin)
	rs.AddLocalFunc("verify_pin", ussdHandlers.VerifyPin)
	rs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	rs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	rs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	rs.AddLocalFunc("quit", ussdHandlers.Quit)
	rs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	rs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	rs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	rs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	rs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	rs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	rs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	rs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	rs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	rs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	rs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	rs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	rs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	rs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	rs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	rs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	rs.AddLocalFunc("quit_with_balance", ussdHandlers.QuitWithBalance)
	rs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	rs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	rs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	rs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	rs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	rs.AddLocalFunc("set_reset_single_edit", ussdHandlers.SetResetSingleEdit)
	rs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)
	rs.AddLocalFunc("save_temporary_pin", ussdHandlers.SaveTemporaryPin)
	rs.AddLocalFunc("verify_new_pin", ussdHandlers.VerifyNewPin)
	rs.AddLocalFunc("confirm_pin_change", ussdHandlers.ConfirmPinChange)
	rs.AddLocalFunc("quit_with_help",ussdHandlers.QuitWithHelp)

	return ussdHandlers, nil
}

func ensureDbDir(dbDir string) error {
	err := os.MkdirAll(dbDir, 0700)
	if err != nil {
		return fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	return nil
}

func getStateStore(dbDir string, ctx context.Context) (db.Db, error) {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	return store, nil
}

func getUserdataDb(dbDir string, ctx context.Context) db.Db {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "userdata.gdbm")
	store.Connect(ctx, storeFile)
	
	return store
}

func getResource(resourceDir string, ctx context.Context) (resource.Resource, error) {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(store)
	return rfs, nil
}


func main() {
	var sessionId string
	var dbDir string
	var resourceDir string
	var size uint
	var engineDebug bool
	var stateDebug bool
	var host string
	var port uint
	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.BoolVar(&engineDebug, "engine-debug", false, "use engine debug output")
	flag.BoolVar(&stateDebug, "state-debug", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", "127.0.0.1", "http host")
	flag.UintVar(&port, "p", 7123, "http port")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "resourcedir", resourceDir,  "outputsize", size, "sessionId", sessionId)

	ctx := context.Background()
	pfp := path.Join(scriptDir, "pp.csv")
	flagParser, err := getFlags(pfp, true)

	if err != nil {
		os.Exit(1)
	}

	cfg := engine.Config{
		Root:       "root",
		OutputSize: uint32(size),
		FlagCount:  uint32(16),
	}
	if stateDebug {
		cfg.StateDebug = true
	}
	if engineDebug {
		cfg.EngineDebug = true
	}

	rs, err := getResource(resourceDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = ensureDbDir(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userdataStore := getUserdataDb(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer userdataStore.Close()

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		os.Exit(1)
	}

	hl, err := getHandler(flagParser, dbResource, userdataStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	stateStore, err := getStateStore(dbDir, ctx)
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
		Ctx: ctx,
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
			fmt.Errorf("error in process: %v", err)
			os.Exit(1)
		}
		rqs, err = sh.Output(rqs)
		if err != nil {
			fmt.Errorf("error in output: %v", err)
			os.Exit(1)
		}
		rqs, err = sh.Reset(rqs)
		if err != nil {
			fmt.Errorf("error in reset: %v", err)
			os.Exit(1)
		}
		fmt.Println("")
		_, err = fmt.Scanln(&rqs.Input)
		if err != nil {
		fmt.Errorf("error in input: %v", err)
			os.Exit(1)
		}
	}
}
