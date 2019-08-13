package storer

import (
	"encoding/hex"
	"errors"

	"github.com/nochso/bytesort"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Update ...
func (list *List) Update(fn func(txn *ListTxn) error) error {
	return list.m.store.Update(func(txn kvstore.Txn) error {
		return fn(list.Txn(txn))
	})
}

// View ...
func (list *List) View(fn func(txn *ListTxn) error) error {
	return list.m.store.View(func(txn kvstore.Txn) error {
		return fn(list.Txn(txn))
	})
}

// Add string version of AddByte
func (t *ListTxn) Add(item interface{}) (string, error) {
	id, err := t.AddByBytes(item)
	return hex.EncodeToString(id), err
}

// Add auto transaction version of ListTxn.Add
func (list *List) Add(item interface{}) (id string, err error) {
	err = list.Update(func(txn *ListTxn) error {
		id, err = txn.Add(item)
		return err
	})
	return
}

// Get auto transaction version of ListTxn.Get
func (list *List) Get(id string, item interface{}) (err error) {
	err = list.View(func(txn *ListTxn) error {
		return txn.Get(id, item)
	})
	return
}

// Set auto transaction version of ListTxn.Set
func (list *List) Set(id string, item interface{}) error {
	return list.Update(func(txn *ListTxn) error {
		return txn.Set(id, item)
	})
}

// Del auto transaction version of ListTxn.Del
func (list *List) Del(id string) error {
	return list.Update(func(txn *ListTxn) error {
		return txn.Del(id)
	})
}

// Get string version of GetByte
func (t *ListTxn) Get(id string, item interface{}) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}
	return t.GetByBytes(b, item)
}

// Set string version of SetByte
func (t *ListTxn) Set(id string, item interface{}) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}

	return t.SetByBytes(b, item)
}

// Del string version of DelByte
func (t *ListTxn) Del(id string) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}

	return t.DelByBytes(b)
}

// Index create index
func (list *List) Index(id string, fn GenIndex) (index *Index) {
	err := list.Update(func(txn *ListTxn) error {
		var err error
		index, err = txn.IndexByBytes(id, func(ctx *GenCtx) ([]byte, error) {
			i := fn(ctx)
			if err, ok := i.(error); ok {
				return nil, err
			}
			return bytesort.Encode(i)
		})
		return err
	})
	if err != nil {
		panic(err)
	}
	return
}

// ErrUniqueIndex ...
var ErrUniqueIndex = errors.New("index already exists")

// UniqueIndex ...
func (list *List) UniqueIndex(id string, fn GenIndex) (index *Index) {
	index = list.Index(id, func(ctx *GenCtx) interface{} {
		i := fn(ctx)
		if err, ok := i.(error); ok {
			return err
		}

		encoded, err := bytesort.Encode(i)
		if err != nil {
			return err
		}

		if ctx.Action != IndexCreate {
			return nil
		}

		has, err := index.Txn(ctx.Txn).FromByBytes(encoded).Has()
		if err != nil {
			return err
		}
		if has {
			return ErrUniqueIndex
		}
		return encoded
	})
	return
}
