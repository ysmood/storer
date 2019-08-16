package storer

import (
	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

// New a shortcut to use badger as the backend.
// The "pkg/kvstore/badger.go" is an example to implement the "kvstore.Store" interface.
func New(dir string) *Store {
	return NewWithDB("", badger.New(dir))
}

// Txn ...
type Txn = kvstore.Txn

// Update ...
func (store *Store) Update(fn kvstore.DoTxn) error {
	return store.db.Do(true, fn)
}

// View ...
func (store *Store) View(fn kvstore.DoTxn) error {
	return store.db.Do(false, fn)
}

// Map create a map
func (store *Store) Map(item interface{}) *Map {
	return store.MapWithName("", item)
}

// List create a list, the name must be unique among all lists
func (store *Store) List(item interface{}) *List {
	return store.ListWithName("", item)
}

// Value create a value store, the item is also the init value
func (store *Store) Value(name string, item interface{}) *Value {
	dict := store.MapWithName(name, item)
	err := dict.GetByBytes(nil, item)
	if err == ErrKeyNotFound {
		utils.E(dict.SetByBytes(nil, item))
	}
	return &Value{dict: dict}
}
