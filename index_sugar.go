package storer

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/nochso/bytesort"
	"github.com/ysmood/storer/pkg/kvstore"
)

// GenIndex if it returns error the transaction will fail
type GenIndex func(*GenCtx) interface{}

// For string version of ForByBytes
func (txnCtx *IndexTxn) For(from interface{}) *ForCtx {
	b, err := bytesort.Encode(from)
	if err != nil {
		panic(err)
	}
	return txnCtx.ForByBytes(b)
}

// Prefix whether the prefix matches the whole key
func (ctx *IterCtx) prefix() bool {
	return bytes.Equal(ctx.IndexBytes(), ctx.forCtx.from)
}

// Has ...
func (ctx *ForCtx) Has() (bool, error) {
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
func (ctx *ForCtx) Find(items interface{}) error {
	listValue := reflect.ValueOf(items).Elem()

	var itemType reflect.Type
	isList := listValue.Kind() == reflect.Slice
	if isList {
		itemType = listValue.Type().Elem()
	} else {
		itemType = listValue.Type()
	}

	noItem := true

	err := ctx.Each(func(ctx *IterCtx) error {
		if !ctx.prefix() {
			return ErrStop
		}
		noItem = false
		item := reflect.New(itemType)
		err := ctx.Item(item.Interface())
		if err != nil {
			return err
		}
		if isList {
			listValue.Set(reflect.Append(listValue, item.Elem()))
		} else {
			listValue.Set(item.Elem())
			return ErrStop
		}
		return nil
	})

	if noItem {
		return ErrNotFound
	}
	return err
}

// ReduceFilter ...
type ReduceFilter func(*IterCtx) (interface{}, error)

// Reduce ...
func (ctx *ForCtx) Reduce(items interface{}, fn ReduceFilter) error {
	listValue := reflect.ValueOf(items).Elem()

	return ctx.Each(func(ctx *IterCtx) error {
		item, err := fn(ctx)
		if err != nil {
			return err
		}

		listValue.Set(reflect.Append(listValue, reflect.ValueOf(item).Elem()))
		return nil
	})
}

// RangeFilter ...
type RangeFilter func(*IterCtx) bool

// Range ...
func (ctx *ForCtx) Range(items interface{}, fn RangeFilter) error {
	listValue := reflect.ValueOf(items).Elem()
	itemType := listValue.Type().Elem()

	return ctx.Reduce(items, func(ctx *IterCtx) (interface{}, error) {
		if !fn(ctx) {
			return nil, ErrStop
		}

		item := reflect.New(itemType)
		err := ctx.Item(item.Interface())
		if err != nil {
			return nil, err
		}
		return item.Interface(), nil
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

// ForTxnCtx ...
type ForTxnCtx struct {
	index *Index
	from  interface{}
}

// For ...
func (index *Index) For(from interface{}) *ForTxnCtx {
	return &ForTxnCtx{
		index: index,
		from:  from,
	}
}

// Find ...
func (ctx *ForTxnCtx) Find(item interface{}) error {
	return ctx.index.list.m.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).For(ctx.from).Find(item)
	})
}

// Range ...
func (ctx *ForTxnCtx) Range(items interface{}, fn RangeFilter) error {
	return ctx.index.list.m.store.View(func(txn kvstore.Txn) error {
		return ctx.index.Txn(txn).For(ctx.from).Range(items, fn)
	})
}
