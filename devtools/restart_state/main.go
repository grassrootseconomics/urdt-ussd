package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	config.LoadConfig()

	var dbDir string
	var sessionId string
	var database string

	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&database, "db", "gdbm", "database to be used")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.Parse()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	ctx = context.WithValue(ctx, "Database", database)

	resourceDir := scriptDir
	menuStorageService := storage.NewMenuStorageService(dbDir, resourceDir)

	err := menuStorageService.EnsureDbDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	pe, err := menuStorageService.GetPersister(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	st := pe.GetState()

	if st == nil {
		logg.ErrorCtxf(ctx, "state fail in devtool", "state", st)
		fmt.Println("cannot get state")
		os.Exit(1)
	}

	fmt.Println("The state:", st)

	// set empty Code to allow the menu to run from the top
	st.Code = []byte{}

	err = pe.Save(sessionId)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to persist the state and cache", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}
