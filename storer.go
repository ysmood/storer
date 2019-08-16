package storer

import (
	"strings"
	"sync"

	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/typee"
)

// ErrKeyNotFound ...
var ErrKeyNotFound = kvstore.ErrKeyNotFound

// Database ...
type Database interface {
	kvstore.Store

	// Close close database file
	Close() error
}

// Store ...
type Store struct {
	// name so that you can use multiple stores for the same database
	name string

	// db the backend for storing data
	db Database

	bucketCache *sync.Map
}

// NewWithDB use your custom backend as the database
func NewWithDB(name string, db Database) *Store {
	return &Store{
		name:        name,
		db:          db,
		bucketCache: &sync.Map{},
	}
}

// Close close database
func (store *Store) Close() error {
	return store.db.Close()
}

// MapWithName ...
func (store *Store) MapWithName(name string, item interface{}) *Map {
	typeID := typee.GenTypeID(item)

	return &Map{
		name:   name,
		store:  store,
		typeID: typeID,
		bucket: store.bucket(typeID.Anchor, name),
	}
}

// ListWithName ...
func (store *Store) ListWithName(name string, item interface{}) *List {
	return &List{
		dict:    store.MapWithName(name, item),
		indexes: map[string]*Index{},
	}
}

// The prefix of the created bucket will be like "mydb:list:users"
func (store *Store) bucket(names ...string) *bucket.Bucket {
	name := strings.Join(append([]string{store.name}, names...), ":")

	var b *bucket.Bucket
	utils.E(store.Update(func(txn kvstore.Txn) error {
		var err error
		b, err = bucket.New(txn, []byte(name))
		return err
	}))

	return b
}
