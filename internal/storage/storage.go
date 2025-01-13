package storage

import (
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/persist"
)

const (
	DATATYPE_EXTEND = 128
)

type Storage struct {
	Persister *persist.Persister
	UserdataDb db.Db	
}

type StorageProvider interface {
	Get(sessionId string) (*Storage, error)
	Put(sessionId string, storage *Storage) error
	Close() error
}

type SimpleStorageProvider struct {
	*Storage
}

func NewSimpleStorageProvider(stateStore db.Db, userdataStore db.Db) StorageProvider {
	pe := persist.NewPersister(stateStore)
	pe = pe.WithFlush()
	return &SimpleStorageProvider{
		Storage: &Storage{
			Persister: pe,
			UserdataDb: userdataStore,
		},
	}
}

func (p *SimpleStorageProvider) Get(sessionId string) (*Storage, error) {
	return p.Storage, nil
}

func (p *SimpleStorageProvider) Put(sessionId string, storage *Storage) error {
	return nil
}

func (p *SimpleStorageProvider) Close() error {
	return p.Storage.UserdataDb.Close()
}
