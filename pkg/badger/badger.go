package badger

import (
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Badger adapter for badger
type Badger struct {
	kvstore.Store

	db *badger.DB
}

// New a helper to create a badger adapter instance.
// If the dir is empty a tmp dir will be created.
func New(dir string) *Badger {
	if dir == "" {
		dir = fmt.Sprintf("tmp/%d", time.Now().UnixNano())
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			panic(err)
		}
	}

	dbOpts := badger.DefaultOptions(dir).WithLogger(nil)
	db, err := badger.Open(dbOpts)
	if err != nil {
		panic(err)
	}

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
	kvstore.Txn

	txn *badger.Txn
}

// Get ...
func (t *Txn) Get(key []byte) ([]byte, error) {
	item, err := t.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, kvstore.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	val, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return val, nil
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
		err := fn(&IterCtx{
			iter: it,
		})
		if err == kvstore.ErrStop {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// IterCtx ...
type IterCtx struct {
	kvstore.IterCtx

	iter *badger.Iterator
}

// Seek ...
func (i *IterCtx) Seek(key []byte) {
	i.iter.Seek(key)
}

// Key ...
func (i *IterCtx) Key() []byte {
	return i.iter.Item().Key()
}
