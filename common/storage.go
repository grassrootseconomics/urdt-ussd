package common

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

func StoreToDb(store *UserDataStore) db.Db {
	return store.Db
}

func StoreToPrefixDb(store *UserDataStore, pfx []byte) storage.PrefixDb {
	return storage.NewSubPrefixDb(store.Db, pfx)	
}

type StorageServices interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) (db.Db, error)
	GetResource(ctx context.Context) (resource.Resource, error)
	EnsureDbDir() error
}

type StorageService struct {
	svc *storage.MenuStorageService
}

func NewStorageService(dbDir string) *StorageService {
	return &StorageService{
		svc: storage.NewMenuStorageService(dbDir, ""),
	}
}

func(ss *StorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	return ss.svc.GetPersister(ctx)
}
	
func(ss *StorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	return ss.svc.GetUserdataDb(ctx)
}

func(ss *StorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	return nil, errors.New("not implemented")
}

func(ss *StorageService) EnsureDbDir() error {
	return ss.svc.EnsureDbDir()
}
