package storage

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/initializers"
)

var (
	logg = logging.NewVanilla().WithDomain("storage")
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) db.Db
	GetResource(ctx context.Context) (resource.Resource, error)
	EnsureDbDir() error
}

type MenuStorageService struct {
	dbDir         string
	resourceDir   string
	resourceStore db.Db
	stateStore    db.Db
	userDataStore db.Db
}

func buildConnStr() string {
	host := initializers.GetEnv("DB_HOST", "localhost")
	user := initializers.GetEnv("DB_USER", "postgres")
	password := initializers.GetEnv("DB_PASSWORD", "")
	dbName := initializers.GetEnv("DB_NAME", "")
	port := initializers.GetEnv("DB_PORT", "5432")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbName,
	)
	logg.Debugf("pg conn string", "conn", connString)

	return connString
}

func NewMenuStorageService(dbDir string, resourceDir string) *MenuStorageService {
	return &MenuStorageService{
		dbDir:       dbDir,
		resourceDir: resourceDir,
	}
}

func (ms *MenuStorageService) getOrCreateDb(ctx context.Context, existingDb db.Db, fileName string) (db.Db, error) {
	database, ok := ctx.Value("Database").(string)
	if !ok {
		return nil, fmt.Errorf("failed to select the database")
	}

	if existingDb != nil {
		return existingDb, nil
	}

	var newDb db.Db
	var err error

	if database == "postgres" {
		newDb = postgres.NewPgDb()
		connStr := buildConnStr()
		err = newDb.Connect(ctx, connStr)
	} else {
		newDb = NewThreadGdbmDb()
		storeFile := path.Join(ms.dbDir, fileName)
		err = newDb.Connect(ctx, storeFile)
	}

	if err != nil {
		return nil, err
	}

	return newDb, nil
}

func (ms *MenuStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	stateStore, err := ms.GetStateStore(ctx)
	if err != nil {
		return nil, err
	}

	pr := persist.NewPersister(stateStore)
	logg.TraceCtxf(ctx, "menu storage service", "persist", pr, "store", stateStore)
	return pr, nil
}

func (ms *MenuStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	if ms.userDataStore != nil {
		return ms.userDataStore, nil
	}

	userDataStore, err := ms.getOrCreateDb(ctx, ms.userDataStore, "userdata.gdbm")
	if err != nil {
		return nil, err
	}

	ms.userDataStore = userDataStore
	return ms.userDataStore, nil
}

func (ms *MenuStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	ms.resourceStore = fsdb.NewFsDb()
	err := ms.resourceStore.Connect(ctx, ms.resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(ms.resourceStore)
	return rfs, nil
}

func (ms *MenuStorageService) GetStateStore(ctx context.Context) (db.Db, error) {
	if ms.stateStore != nil {
		return ms.stateStore, nil
	}

	stateStore, err := ms.getOrCreateDb(ctx, ms.stateStore, "state.gdbm")
	if err != nil {
		return nil, err
	}

	ms.stateStore = stateStore
	return ms.stateStore, nil
}

func (ms *MenuStorageService) EnsureDbDir() error {
	err := os.MkdirAll(ms.dbDir, 0700)
	if err != nil {
		return fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	return nil
}

func (ms *MenuStorageService) Close() error {
	errA := ms.stateStore.Close()
	errB := ms.userDataStore.Close()
	errC := ms.resourceStore.Close()
	if errA != nil || errB != nil || errC != nil {
		return fmt.Errorf("%v %v %v", errA, errB, errC)
	}
	return nil
}
