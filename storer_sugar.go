package storer

import (
	"github.com/ysmood/storer/pkg/kvstore"
)

// Txn ...
type Txn = kvstore.Txn

// Update ...
func (store *Store) Update(fn kvstore.DoTxn) error {
	return store.DB.Do(true, fn)
}

// View ...
func (store *Store) View(fn kvstore.DoTxn) error {
	return store.DB.Do(false, fn)
}
