package common

import (
	"context"

	"git.defalsify.org/vise.git/db"
)

type DataStore interface {
	db.Db
	ReadEntry(ctx context.Context, sessionId string, typ DataTyp) ([]byte, error)
	WriteEntry(ctx context.Context, sessionId string, typ DataTyp, value []byte) error
}

type UserDataStore struct {
	db.Db
}

// ReadEntry retrieves an entry to the userdata store.
func (store *UserDataStore) ReadEntry(ctx context.Context, sessionId string, typ DataTyp) ([]byte, error) {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := PackKey(typ, []byte(sessionId))
	return store.Get(ctx, k)
}

// WriteEntry adds an entry to the userdata store.
// BUG: this uses sessionId twice
func (store *UserDataStore) WriteEntry(ctx context.Context, sessionId string, typ DataTyp, value []byte) error {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := PackKey(typ, []byte(sessionId))
	return store.Put(ctx, k, value)
}
