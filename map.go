package storer

import (
	"errors"
	"reflect"

	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

// ErrItemType ...
var ErrItemType = errors.New("wrong item type")

// Map the base type of the system
type Map struct {
	name     string
	store    *Store
	itemType reflect.Type
	bucket   *bucket.Bucket
}

// MapTxn ...
type MapTxn struct {
	dict *Map
	txn  kvstore.Txn
}

// Txn create transaction context
func (dict *Map) Txn(txn kvstore.Txn) *MapTxn {
	return &MapTxn{
		dict: dict,
		txn:  txn,
	}
}

// SetByBytes set an item to the map
func (dict *MapTxn) SetByBytes(id []byte, item interface{}) error {
	if dict.dict.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	data, err := Encode(item)
	if err != nil {
		return err
	}

	return dict.txn.Set(dict.dict.bucket.Prefix(id), data)
}

// GetByBytes get item from the map
func (dict *MapTxn) GetByBytes(id []byte, item interface{}) error {
	if dict.dict.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	raw, err := dict.txn.Get(dict.dict.bucket.Prefix(id))
	if err != nil {
		return err
	}
	return Decode(raw, item)
}

// DelByBytes remove a item from the map
func (dict *MapTxn) DelByBytes(id []byte) error {
	return dict.txn.Delete(dict.dict.bucket.Prefix(id))
}
