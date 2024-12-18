package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/debug"
	"git.defalsify.org/vise.git/logging"
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
	var engineDebug bool

	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&database, "db", "gdbm", "database to be used")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.Parse()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	ctx = context.WithValue(ctx, "Database", database)

	resourceDir := scriptDir
	menuStorageService := storage.NewMenuStorageService(dbDir, resourceDir)

	store, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	d, err := store.Dump(ctx, []byte(sessionId))
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	for true {
		k, v := d.Next(ctx)
		if k == nil {
			break
		}
		o, err := debug.FromKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Printf("%vValue: %v\n\n", o, v)
	}

	err = store.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
