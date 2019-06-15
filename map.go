package storer

import (
	"reflect"

	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Map the base type of the system
type Map struct {
	name     string
	store    *Store
	itemType reflect.Type
	bucket   *bucket.Bucket
}

// MapTxn ...
type MapTxn struct {
	m   *Map
	txn kvstore.Txn
}

// Txn create transaction context
func (m *Map) Txn(txn kvstore.Txn) *MapTxn {
	return &MapTxn{
		m:   m,
		txn: txn,
	}
}

// SetByBytes set an item to the map
func (t *MapTxn) SetByBytes(id []byte, item interface{}) error {
	if t.m.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	data, err := Encode(item)
	if err != nil {
		return err
	}

	return t.txn.Set(t.m.bucket.Prefix(id), data)
}

// GetByBytes get item from the map
func (t *MapTxn) GetByBytes(id []byte, item interface{}) error {
	raw, err := t.txn.Get(t.m.bucket.Prefix(id))
	if err != nil {
		return err
	}
	return Decode(raw, item)
}

// DelByBytes remove a item from the map
func (t *MapTxn) DelByBytes(id []byte) error {
	return t.txn.Delete(t.m.bucket.Prefix(id))
}
