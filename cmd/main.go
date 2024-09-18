package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

func getParser(fp string, debug bool) (*asm.FlagParser, error) {
	flagParser := asm.NewFlagParser().WithDebug()
	_, err := flagParser.Load(fp)
	if err != nil {
		return nil, err
	}
	return flagParser, nil
}

func getHandler(appFlags *asm.FlagParser, rs *resource.DbResource, pe *persist.Persister, userdataStore db.Db) (*ussd.Handlers, error) {

	ussdHandlers, err := ussd.NewHandlers(appFlags, userdataStore)
	if err != nil {
		return nil, err
	}
	ussdHandlers = ussdHandlers.WithPersister(pe)
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

func getPersister(dbDir string, ctx context.Context) (*persist.Persister, error) {
	err := ensureDbDir(dbDir)
	if err != nil {
		return nil, err
	}
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	pr := persist.NewPersister(store)
	return pr, nil
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

func getEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) *engine.DefaultEngine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func main() {
	var dbDir string
	var size uint
	var sessionId string
	var debug bool
	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.BoolVar(&debug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "outputsize", size)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	pfp := path.Join(scriptDir, "pp.csv")
	flagParser, err := getParser(pfp, true)

	if err != nil {
		os.Exit(1)
	}

	cfg := engine.Config{
		Root:       "root",
		SessionId:  sessionId,
		OutputSize: uint32(size),
		FlagCount:  uint32(16),
	}

	rs, err := getResource(scriptDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	pe, err := getPersister(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	store := getUserdataDb(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	hl, err := getHandler(flagParser, dbResource, pe, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	en := getEngine(cfg, rs, pe)
	en = en.WithFirst(hl.Init)
	if debug {
		en = en.WithDebug(nil)
	}

	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
