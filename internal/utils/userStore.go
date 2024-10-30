package utils

import (
	"context"

	"git.grassecon.net/urdt/ussd/common"

	"git.defalsify.org/vise.git/db"
)

type DataStore interface {
	db.Db
	ReadEntry(ctx context.Context, sessionId string, typ common.DataTyp) ([]byte, error)
	WriteEntry(ctx context.Context, sessionId string, typ common.DataTyp, value []byte) error
}

type UserDataStore struct {
	db.Db
}

// ReadEntry retrieves an entry from the store based on the provided parameters.
func (store *UserDataStore) ReadEntry(ctx context.Context, sessionId string, typ common.DataTyp) ([]byte, error) {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := common.PackKey(typ, []byte(sessionId))
	return store.Get(ctx, k)
}

func (store *UserDataStore) WriteEntry(ctx context.Context, sessionId string, typ common.DataTyp, value []byte) error {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := common.PackKey(typ, []byte(sessionId))
	return store.Put(ctx, k, value)
}
