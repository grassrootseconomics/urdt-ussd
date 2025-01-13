package main

import (
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/urdt/ussd/common"
	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/storage"
	testdataloader "github.com/peteole/testdata-loader"
)

var (
	logg      = logging.NewVanilla()
	baseDir   = testdataloader.GetBasePath()
	scriptDir = path.Join("services", "registration")
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	config.LoadConfig()

	var connStr string
	var sessionId string
	var database string
	var engineDebug bool
	var err error

	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&connStr, "c", "", "connection string")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.Parse()

	if connStr != "" {
		connStr = config.DbConn
	}
	connData, err := storage.ToConnData(config.DbConn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connstr err: %v", err)
		os.Exit(1)
	}

	logg.Infof("start command", "conn", connData)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	ctx = context.WithValue(ctx, "Database", database)

	resourceDir := scriptDir
	menuStorageService := storage.NewMenuStorageService(connData, resourceDir)
	
	store, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	userStore := common.UserDataStore{store}

	h := sha1.New()
	h.Write([]byte(sessionId))
	address := h.Sum(nil)
	addressString := fmt.Sprintf("%x", address)

	err = userStore.WriteEntry(ctx, sessionId, common.DATA_PUBLIC_KEY, []byte(addressString))
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = userStore.WriteEntry(ctx, addressString, common.DATA_PUBLIC_KEY_REVERSE, []byte(sessionId))
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = store.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
