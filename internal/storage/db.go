package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
)

const (
	DATATYPE_USERSUB = 64
)

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
	v, err := s.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *SubPrefixDb) Put(ctx context.Context, key []byte, val []byte) error {
	s.store.SetPrefix(DATATYPE_USERSUB)
	key = s.toKey(key)
	return s.store.Put(ctx, key, val)
}
