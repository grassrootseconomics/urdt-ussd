package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
)

const (
	DATATYPE_USERSUB = 64
)

// PrefixDb interface abstracts the database operations.
type PrefixDb interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, key []byte, val []byte) error
}

var _ PrefixDb = (*SubPrefixDb)(nil)

type SubPrefixDb struct {
	store db.Db
	pfx   []byte
}

func NewSubPrefixDb(store db.Db, pfx []byte) *SubPrefixDb {
	return &SubPrefixDb{
		store: store,
		pfx:   pfx,
	}
}

func (s *SubPrefixDb) toKey(k []byte) []byte {
	return append(s.pfx, k...)
}

func (s *SubPrefixDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	s.store.SetPrefix(DATATYPE_USERSUB)
	key = s.toKey(key)
	return s.store.Get(ctx, key)
}

func (s *SubPrefixDb) Put(ctx context.Context, key []byte, val []byte) error {
	s.store.SetPrefix(DATATYPE_USERSUB)
	key = s.toKey(key)
	return s.store.Put(ctx, key, val)
}
