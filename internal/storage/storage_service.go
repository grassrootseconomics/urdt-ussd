package storage

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	gdbmstorage "git.grassecon.net/urdt/ussd/internal/storage/db/gdbm"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	logg = logging.NewVanilla().WithDomain("storage")
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) db.Db
	GetResource(ctx context.Context) (resource.Resource, error)
}

type MenuStorageService struct {
	conn ConnData
	resourceDir   string
	poResource    resource.Resource
	resourceStore db.Db
	stateStore    db.Db
	userDataStore db.Db
}

func NewMenuStorageService(conn ConnData, resourceDir string) *MenuStorageService {
	return &MenuStorageService{
		conn: conn,
		resourceDir: resourceDir,
	}
}

func (ms *MenuStorageService) getOrCreateDb(ctx context.Context, existingDb db.Db, section string) (db.Db, error) {
	var newDb db.Db
	var err error

	if existingDb != nil {
		return existingDb, nil
	}


	connStr := ms.conn.String()
	dbTyp := ms.conn.DbType()
	if dbTyp == DBTYPE_POSTGRES {
		// TODO: move to vise
		err = ensureSchemaExists(ctx, ms.conn)
		if err != nil {
			return nil, err
		}
		newDb = postgres.NewPgDb().WithSchema(ms.conn.Domain())
	} else if dbTyp == DBTYPE_GDBM {
		err = ms.ensureDbDir()
		if err != nil {
			return nil, err
		}
		connStr = path.Join(connStr, section)
		newDb = gdbmstorage.NewThreadGdbmDb()
	} else {
		return nil, fmt.Errorf("unsupported connection string: '%s'\n", ms.conn.String())
	}
	logg.DebugCtxf(ctx, "connecting to db", "conn", connStr, "conndata", ms.conn)
	err = newDb.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return newDb, nil
}

// WithGettext triggers use of gettext for translation of templates and menus.
//
// The first language in `lns` will be used as default language, to resolve node keys to 
// language strings.
//
// If `lns` is an empty array, gettext will not be used.
func (ms *MenuStorageService) WithGettext(path string, lns []lang.Language) *MenuStorageService {
	if len(lns) == 0 {
		logg.Warnf("Gettext requested but no languages supplied")
		return ms
	}
	rs := resource.NewPoResource(lns[0], path)

	for _, ln := range(lns) {
		rs = rs.WithLanguage(ln)
	}

	ms.poResource = rs

	return ms
}

// ensureSchemaExists creates a new schema if it does not exist
func ensureSchemaExists(ctx context.Context, conn ConnData) error {
	h, err := pgxpool.New(ctx, conn.Path())
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer h.Close()

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", conn.Domain())
	_, err = h.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
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
	if ms.poResource != nil {
		logg.InfoCtxf(ctx, "using poresource for menu and template")
		rfs.WithMenuGetter(ms.poResource.GetMenu)
		rfs.WithTemplateGetter(ms.poResource.GetTemplate)
	}
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

func (ms *MenuStorageService) ensureDbDir() error {
	err := os.MkdirAll(ms.conn.String(), 0700)
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
