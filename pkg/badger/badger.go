package badger

import (
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger"
	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Badger adapter for badger
type Badger struct {
	db *badger.DB
}

var _ kvstore.Store = &Badger{}

// New a helper to create a badger adapter instance.
// If the dir is empty a tmp dir will be created.
func New(dir string) *Badger {
	if dir == "" {
		dir = filepath.Join("tmp", utils.RandString(10))
	}

	err := os.MkdirAll(dir, 0775)
	utils.E(err)

	dbOpts := badger.DefaultOptions(dir).WithLogger(nil)
	db, err := badger.Open(dbOpts)
	utils.E(err)

	return NewByDB(db)
}

// NewByDB ...
func NewByDB(db *badger.DB) *Badger {
	return &Badger{
		db: db,
	}
}

// Do ...
func (b *Badger) Do(update bool, fn kvstore.DoTxn) error {
	txn := b.db.NewTransaction(update)
	defer txn.Discard()

	err := fn(&Txn{
		txn: txn,
	})
	if err != nil {
		return err
	}

	if update {
		return txn.Commit()
	}
	return nil
}

// Close ...
func (b *Badger) Close() error {
	return b.db.Close()
}

// Txn ...
type Txn struct {
	txn *badger.Txn
}

var _ kvstore.Txn = &Txn{}

// Get ...
func (t *Txn) Get(key []byte) ([]byte, error) {
	item, err := t.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, kvstore.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return item.ValueCopy(nil)
}

// Set ...
func (t *Txn) Set(key, value []byte) error {
	return t.txn.Set(key, value)
}

// Delete ...
func (t *Txn) Delete(key []byte) error {
	return t.txn.Delete(key)
}

// Do ...
func (t *Txn) Do(reverse bool, from []byte, fn kvstore.Iteratee) error {
	opts := badger.IteratorOptions{
		PrefetchValues: false,
		PrefetchSize:   0,
		Reverse:        reverse,
		AllVersions:    false,
	}

	it := t.txn.NewIterator(opts)
	defer it.Close()

	for it.Seek(from); it.Valid(); it.Next() {
		err := fn(it.Item().Key())
		if err == kvstore.ErrStop {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}
