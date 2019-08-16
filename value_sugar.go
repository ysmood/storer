package storer

import (
	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/typee"
)

// Value single value store
type Value struct {
	dict *Map
}

// Txn ...
func (v *Value) Txn(txn kvstore.Txn) *ValueTxn {
	return &ValueTxn{
		dictTxn: v.dict.Txn(txn),
	}
}

// ValueTxn ...
type ValueTxn struct {
	dictTxn *MapTxn
}

// Set ...
func (t *ValueTxn) Set(item interface{}) error {
	return t.dictTxn.SetByBytes(nil, item)
}

// Get ...
func (t *ValueTxn) Get(item interface{}) error {
	err := t.dictTxn.GetByBytes(nil, item)
	if err == typee.ErrMigrated {
		return nil
	}
	return err
}

// Set ...
func (v *Value) Set(item interface{}) error {
	return v.dict.SetByBytes(nil, item)
}

// Get ...
func (v *Value) Get(item interface{}) error {
	return v.dict.GetByBytes(nil, item)
}
