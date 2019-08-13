package storer

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/nochso/bytesort"
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
	if err != nil {
		panic(err)
	}
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
var ErrNotFound = errors.New("Not found")

// Find items can be a list or a singular
func (ctx *FromCtx) Find(items interface{}) error {
	listValue := reflect.ValueOf(items).Elem()
	isList := listValue.Kind() == reflect.Slice
	itemType := ctx.txnCtx.index.list.m.itemType.Elem()
	noItem := true

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

	if noItem {
		return ErrNotFound
	}
	return err
}

// Filter ...
type Filter func(ctx *IterCtx) (match bool, isContinue bool)

// Filter ...
func (ctx *FromCtx) Filter(items interface{}, fn Filter) error {
	listValue := reflect.ValueOf(items).Elem()
	itemType := ctx.txnCtx.index.list.m.itemType.Elem()

	return ctx.Each(func(ctx *IterCtx) error {
		match, isContinue := fn(ctx)

		if match {
			item := reflect.New(itemType)
			err := ctx.Item(item.Interface())
			if err != nil {
				return err
			}

			listValue.Set(reflect.Append(listValue, item.Elem()))
		}
		if isContinue {
			return nil
		}

		return ErrStop
	})
}

// ID ...
func (ctx *IterCtx) ID() string {
	return string(ctx.IDBytes())
}

// Index ...
func (ctx *IterCtx) Index() string {
	return string(ctx.IndexBytes())
}

// Compare ...
func (ctx *IterCtx) Compare(v interface{}) int {
	b, err := bytesort.Encode(v)
	if err != nil {
		panic(err)
	}
	return bytes.Compare(ctx.IndexBytes(), b)
}

// Seek ...
func (ctx *IterCtx) Seek(from string) {
	ctx.SeekBytes([]byte(from))
}

// Item ...
func (ctx *IterCtx) Item(item interface{}) error {
	if ctx.forCtx.txnCtx.index.list.m.itemType != reflect.TypeOf(item) {
		return ErrItemType
	}

	return ctx.forCtx.txnCtx.index.list.Txn(
		ctx.forCtx.txnCtx.txn,
	).GetByBytes(ctx.IDBytes(), item)
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
	return ctx.index.list.m.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Each(fn)
	})
}

// Find ...
func (ctx *FromTxnCtx) Find(item interface{}) error {
	return ctx.index.list.m.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Find(item)
	})
}

// Filter ...
func (ctx *FromTxnCtx) Filter(items interface{}, fn Filter) error {
	return ctx.index.list.m.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).From(ctx.from).Filter(items, fn)
	})
}
