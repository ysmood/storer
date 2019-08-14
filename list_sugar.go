package storer

import (
	"encoding/hex"
	"errors"
	"reflect"

	"github.com/nochso/bytesort"
	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Update ...
func (list *List) Update(fn func(txn *ListTxn) error) error {
	return list.dict.store.Update(func(txn kvstore.Txn) error {
		return fn(list.Txn(txn))
	})
}

// View ...
func (list *List) View(fn func(txn *ListTxn) error) error {
	return list.dict.store.View(func(txn kvstore.Txn) error {
		return fn(list.Txn(txn))
	})
}

// Add string version of AddByte
func (listTxn *ListTxn) Add(item interface{}) (string, error) {
	id, err := listTxn.AddByBytes(item)
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
func (listTxn *ListTxn) Get(id string, item interface{}) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}
	return listTxn.GetByBytes(b, item)
}

// Set string version of SetByte
func (listTxn *ListTxn) Set(id string, item interface{}) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}

	return listTxn.SetByBytes(b, item)
}

// Del string version of DelByte
func (listTxn *ListTxn) Del(id string) error {
	b, err := hex.DecodeString(id)
	if err != nil {
		return err
	}

	return listTxn.DelByBytes(b)
}

func (list *List) indexCallback(fn interface{}) GenIndex {
	cb, ok := fn.(func(ctx *GenCtx) interface{})
	if !ok {
		cb = func(ctx *GenCtx) interface{} {
			return reflect.ValueOf(fn).Call(
				[]reflect.Value{reflect.ValueOf(ctx.Item)},
			)[0].Interface()
		}
	}
	return cb
}

// Index create index, fn can be GenIndex
func (list *List) Index(id string, fn interface{}) (index *Index) {
	cb := list.indexCallback(fn)

	err := list.Update(func(txn *ListTxn) error {
		var err error
		index, err = txn.IndexByBytes(id, func(ctx *GenCtx) ([]byte, error) {
			i := cb(ctx)
			if err, ok := i.(error); ok {
				return nil, err
			}
			return bytesort.Encode(i)
		})
		return err
	})
	utils.E(err)
	return
}

// ErrUniqueIndex ...
var ErrUniqueIndex = errors.New("index already exists")

// UniqueIndex ...
func (list *List) UniqueIndex(id string, fn interface{}) (index *Index) {
	cb := list.indexCallback(fn)

	index = list.Index(id, func(ctx *GenCtx) interface{} {
		i := cb(ctx)
		if err, ok := i.(error); ok {
			return err
		}

		if ctx.Action != IndexCreate {
			return i
		}

		has, err := index.Txn(ctx.Txn).From(i).Has()

		if err != nil {
			return err
		}
		if has {
			return ErrUniqueIndex
		}
		return i
	})
	return
}
