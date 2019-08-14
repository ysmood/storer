package storer

import (
	"errors"
	"reflect"
	"strings"

	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Database ...
type Database interface {
	kvstore.Store

	// Close close database file
	Close() error
}

// Store ...
type Store struct {
	// Name so that you can use multiple stores for the same database
	Name string

	// DB the backend for storing data
	DB Database
}

// New a shortcut to use badger as the backend.
// The "pkg/kvstore/badger.go" is an example to implement the "kvstore.Store" interface.
func New(dir string) *Store {
	return &Store{
		DB: badger.New(dir),
	}
}

// Close close database
func (store *Store) Close() error {
	return store.DB.Close()
}

// ErrItemPtr when the item argument is no a pointer
var ErrItemPtr = errors.New("must be a pointer to the item")

// create a new collection
func (store *Store) new(dataType string, item interface{}) *Map {
	t, name := genTypeID(item)

	if t.Kind() != reflect.Ptr {
		panic(ErrItemPtr)
	}

	bucket, err := store.bucket(dataType, name)
	utils.E(err)

	return &Map{
		name:     name,
		store:    store,
		itemType: t,
		bucket:   bucket,
	}
}

// Map create a map
func (store *Store) Map(item interface{}) *Map {
	return store.new(mapType, item)
}

// List create a list, the name must be unique among all lists
func (store *Store) List(item interface{}) *List {
	return &List{
		dict:    store.new(listType, item),
		indexes: map[string]*Index{},
	}
}

// The prefix of the created bucket will be like "mydb:list:users"
func (store *Store) bucket(names ...string) (*bucket.Bucket, error) {
	name := strings.Join(append([]string{store.Name}, names...), ":")

	var b *bucket.Bucket
	err := store.DB.Do(true, func(txn kvstore.Txn) error {
		var err error
		b, err = bucket.New(txn, name)
		return err
	})
	return b, err
}
