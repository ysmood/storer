package storer

import (
	"errors"
	"reflect"

	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/typee"
)

// ErrItemType ...
var ErrItemType = errors.New("[storer] wrong item type")

// Map the base type of the system, such as List is based on Map
// It provides basic set and get object to the database.
type Map struct {
	name   string
	store  *Store
	typeID *typee.TypeID
	bucket *bucket.Bucket
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
	if dictTxn.dict.typeID.Type != reflect.TypeOf(item).Elem() {
		return ErrItemType
	}

	data, err := typee.Encode(item, dictTxn.typeIDMapper)
	if err != nil {
		return err
	}

	return dictTxn.txn.Set(dictTxn.dict.bucket.Prefix(id), data)
}

// GetByBytes get item from the map
func (dictTxn *MapTxn) GetByBytes(id []byte, item interface{}) error {
	if dictTxn.dict.typeID.Type != reflect.TypeOf(item).Elem() {
		return ErrItemType
	}

	raw, err := dictTxn.txn.Get(dictTxn.dict.bucket.Prefix(id))
	if err != nil {
		return err
	}
	err = typee.Decode(raw, item, dictTxn.typeIDMapper)
	if err == typee.ErrMigrated {
		// so that same migration won't happen again
		err = dictTxn.dict.SetByBytes(id, item)
		if err != nil {
			return err
		}
		// upper data structure should also handle this
		return typee.ErrMigrated
	}
	return err
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

func (dictTxn *MapTxn) typeIDMapper(longID []byte) ([]byte, error) {
	if b, ok := dictTxn.dict.store.bucketCache.Load(string(longID)); ok {
		return b.([]byte), nil
	}

	var b *bucket.Bucket
	var err error
	err = dictTxn.dict.store.Update(func(txn Txn) error {
		b, err = bucket.New(txn, longID)
		return err
	})
	if err != nil {
		return nil, err
	}
	shortID := b.Prefix(nil)
	dictTxn.dict.store.bucketCache.Store(string(longID), shortID)
	return shortID, nil
}
