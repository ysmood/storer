package storer

import (
	"errors"
	"reflect"

	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

// ErrItemType ...
var ErrItemType = errors.New("wrong item type")

// Map the base type of the system, such as List is based on Map
// It provides basic set and get object to the database.
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
func (dictTxn *MapTxn) SetByBytes(id []byte, item interface{}) error {
	if dictTxn.dict.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	data, err := Encode(item)
	if err != nil {
		return err
	}

	return dictTxn.txn.Set(dictTxn.dict.bucket.Prefix(id), data)
}

// GetByBytes get item from the map
func (dictTxn *MapTxn) GetByBytes(id []byte, item interface{}) error {
	if dictTxn.dict.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	raw, err := dictTxn.txn.Get(dictTxn.dict.bucket.Prefix(id))
	if err != nil {
		return err
	}
	return Decode(raw, item)
}

// DelByBytes remove a item from the map
func (dictTxn *MapTxn) DelByBytes(id []byte) error {
	return dictTxn.txn.Delete(dictTxn.dict.bucket.Prefix(id))
}

// MapEach ...
type MapEach func(id []byte) error

// Each ...
func (dictTxn *MapTxn) Each(fn MapEach) error {
	return dictTxn.txn.Do(false, dictTxn.dict.bucket.Prefix(nil), func(key []byte) error {
		if !dictTxn.dict.bucket.Valid(key) {
			return ErrStop
		}

		return fn(key[dictTxn.dict.bucket.Len():])
	})
}
