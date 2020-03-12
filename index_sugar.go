package storer

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/nochso/bytesort"
	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/kvstore"
)

// GenIndex if it returns error the transaction will fail
type GenIndex func(ctx *GenCtx) interface{}

// From string version of FromByBytes
func (txnCtx *IndexTxn) From(from interface{}) *FromCtx {
	if from == nil {
		return txnCtx.FromByBytes(nil)
	}

	b, err := bytesort.Encode(from)
	utils.E(err)
	return txnCtx.FromByBytes(b)
}

// Prefix whether the prefix matches the whole key
func (ctx *IterCtx) prefix() bool {
	return bytes.Equal(ctx.IndexBytes(), ctx.forCtx.from)
}

// Has ...
func (ctx *FromCtx) Has() (bool, error) {
	has := false
	err := ctx.Each(func(ctx *IterCtx) error {
		has = ctx.prefix()
		return ErrStop
	})
	return has, err
}

// ErrNotFound ...
var ErrNotFound = errors.New("[storer] not found")

// ErrNoReverse ...
var ErrNoReverse = errors.New("[storer] reverse option is illegal")

// Find items can be a list or a singular
func (ctx *FromCtx) Find(items interface{}) error {
	listValue := reflect.ValueOf(items).Elem()
	isList := listValue.Kind() == reflect.Slice
	itemType := ctx.txnCtx.index.list.dict.typeID.Type
	noItem := true

	// reverse find is meaningless
	if ctx.reverse {
		return ErrNoReverse
	}

	err := ctx.Each(func(ctx *IterCtx) error {
		if !ctx.prefix() {
			return ErrStop
		}
		noItem = false
		if isList {
			item := reflect.New(itemType)
			err := ctx.Item(item.Interface())
			if err != nil {
				return err
			}
			listValue.Set(reflect.Append(listValue, item.Elem()))
		} else {
			err := ctx.Item(items)
			if err != nil {
				return err
			}
			return ErrStop
		}
		return nil
	})
	if err != nil {
		return err
	}
	if noItem {
		return ErrNotFound
	}
	return nil
}

// ErrFilterReturn ...
var ErrFilterReturn = errors.New("[storer] filter wrong return type")

// Filter to stop the filter return ErrStop
func (ctx *FromCtx) Filter(items interface{}, fn interface{}) error {
	listValue := reflect.ValueOf(items).Elem()
	itemType := ctx.txnCtx.index.list.dict.typeID.Type

	return ctx.Each(func(ctx *IterCtx) error {
		item := reflect.New(itemType)
		err := ctx.Item(item.Interface())
		if err != nil {
			return err
		}

		res := reflect.ValueOf(fn).Call(
			[]reflect.Value{item},
		)[0].Interface()

		switch v := res.(type) {
		case bool:
			if v {
				listValue.Set(reflect.Append(listValue, item.Elem()))
			}
			return nil
		case error:
			return v
		default:
			return ErrFilterReturn
		}
	})
}

// Compare ...
func (ctx *IterCtx) Compare(v interface{}) int {
	b, err := bytesort.Encode(v)
	utils.E(err)
	return bytes.Compare(ctx.IndexBytes(), b)
}

// Item ...
func (ctx *IterCtx) Item(item interface{}) error {
	return ctx.forCtx.txnCtx.index.list.Txn(
		ctx.forCtx.txnCtx.txn,
	).GetByBytes(ctx.IDBytes(), item)
}

// Reindex ...
func (index *Index) Reindex() error {
	return index.list.dict.store.Update(func(txn Txn) error {
		return index.Txn(txn).Reindex()
	})
}

// FromTxnCtx ...
type FromTxnCtx struct {
	index *Index
	from  interface{}
}

// From ...
func (index *Index) From(from interface{}) *FromTxnCtx {
	return &FromTxnCtx{
		index: index,
		from:  from,
	}
}

// Each ...
func (ctx *FromTxnCtx) Each(fn Iteratee) error {
	return ctx.index.list.dict.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Each(fn)
	})
}

// Find ...
func (ctx *FromTxnCtx) Find(item interface{}) error {
	return ctx.index.list.dict.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Find(item)
	})
}

// Filter ...
func (ctx *FromTxnCtx) Filter(items interface{}, fn interface{}) error {
	return ctx.index.list.dict.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Filter(items, fn)
	})
}
