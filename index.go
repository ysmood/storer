package storer

import (
	"bytes"
	"reflect"

	"github.com/ysmood/byframe"
	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

// IndexAction ...
type IndexAction int

const (
	// IndexCreate ...
	IndexCreate IndexAction = iota
	// IndexUpdate ...
	IndexUpdate
	// IndexDelete ...
	IndexDelete
)

// GenCtx ...
type GenCtx struct {
	// Item the current item pointer
	Item interface{}
	// Txn ...
	Txn kvstore.Txn
	// Action ...
	Action IndexAction
}

// GenIndexBytes ...
type GenIndexBytes func(*GenCtx) ([]byte, error)

// Index ...
type Index struct {
	name     string
	list     *List
	bucket   *bucket.Bucket
	rbucket  *bucket.Bucket
	genIndex GenIndexBytes
}

// Generate an unique id, format "bucket indexLength index itemID"
// Why use key only for indexing is because we need to make sure
// different items can has the same index.
// The indexLength will make sure only the identical index prefix can be compaired.
func (index *Index) id(i, itemID []byte) []byte {
	return append(index.bucket.Prefix(byframe.Encode(i)), itemID...)
}

// the index part of the key
func (index *Index) extractIndex(key []byte) []byte {
	l := index.bucket.Len()
	id, _, _ := byframe.Decode(key[l:])
	return id
}

// the id of the item
func (index *Index) extractItemID(key []byte) []byte {
	l := index.bucket.Len()
	hLen, indexLen, _ := byframe.DecodeHeader(key[l:])
	return key[l+hLen+indexLen:]
}

func (index *Index) add(txn kvstore.Txn, itemID []byte, item interface{}) error {
	i, err := index.genIndex(&GenCtx{Item: item, Txn: txn, Action: IndexCreate})
	if err != nil {
		return err
	}
	key := index.id(i, itemID)

	err = txn.Set(index.rbucket.Prefix(itemID), i)
	if err != nil {
		return err
	}

	return txn.Set(key, nil)
}

func (index *Index) update(txn kvstore.Txn, itemID []byte, item interface{}) error {
	rid := index.rbucket.Prefix(itemID)
	old, err := txn.Get(rid)
	if err != nil {
		return err
	}

	i, err := index.genIndex(&GenCtx{Item: item, Txn: txn, Action: IndexUpdate})
	if err != nil {
		return err
	}

	if bytes.Equal(old, i) {
		return nil
	}

	err = txn.Set(rid, i)
	if err != nil {
		return err
	}

	return txn.Set(index.id(i, itemID), nil)
}

func (index *Index) del(txn kvstore.Txn, itemID []byte) error {
	rid := index.rbucket.Prefix(itemID)
	i, err := txn.Get(rid)
	if err != nil {
		return err
	}
	err = txn.Delete(rid)
	if err != nil {
		return err
	}
	return txn.Delete(index.id(i, itemID))
}

// IndexTxn ...
type IndexTxn struct {
	index *Index
	txn   kvstore.Txn
}

// Txn ...
func (index *Index) Txn(txn kvstore.Txn) *IndexTxn {
	return &IndexTxn{
		index: index,
		txn:   txn,
	}
}

// FromByBytes ...
func (txnCtx *IndexTxn) FromByBytes(from []byte) *FromCtx {
	return &FromCtx{
		txnCtx: txnCtx,
		from:   from,
	}
}

// Reindex ...
func (txnCtx *IndexTxn) Reindex() error {
	prefix := txnCtx.index.rbucket.Prefix(nil)
	l := len(prefix)
	return txnCtx.txn.Do(false, prefix, func(key []byte) error {
		if !bytes.HasPrefix(key, prefix) {
			return ErrStop
		}
		itemID := key[l:]

		item := reflect.New(txnCtx.index.list.dict.typeID.Type).Interface()
		err := txnCtx.index.list.Txn(txnCtx.txn).GetByBytes(itemID, item)
		if err != nil {
			return err
		}
		return txnCtx.index.update(txnCtx.txn, itemID, item)
	})
}

// FromCtx ...
type FromCtx struct {
	txnCtx  *IndexTxn
	from    []byte
	reverse bool
}

// Reverse iterate reversely
func (ctx *FromCtx) Reverse() *FromCtx {
	ctx.reverse = true
	return ctx
}

// Each ...
func (ctx *FromCtx) Each(fn Iteratee) error {
	var from []byte
	if ctx.from == nil {
		from = []byte{}
	} else {
		from = byframe.Encode(ctx.from)
	}
	prefix := ctx.txnCtx.index.bucket.Prefix(from)

	return ctx.txnCtx.txn.Do(ctx.reverse, prefix, func(key []byte) error {
		// if the key doesn't match the bucket prefix, it means
		// the bucket range is ended, the iteration should stop
		if !ctx.txnCtx.index.bucket.Valid(key) {
			return ErrStop
		}

		return fn(&IterCtx{
			forCtx: ctx,
			key:    key,
		})
	})
}

// IterCtx ...
type IterCtx struct {
	forCtx *FromCtx
	key    []byte
}

// ErrStop ...
var ErrStop = kvstore.ErrStop

// Iteratee params are index and data.
// If returns ErrStop, the iteration will stop
type Iteratee func(*IterCtx) error

// IDBytes ...
func (ctx *IterCtx) IDBytes() []byte {
	return ctx.forCtx.txnCtx.index.extractItemID(ctx.key)
}

// IndexBytes ...
func (ctx *IterCtx) IndexBytes() []byte {
	return ctx.forCtx.txnCtx.index.extractIndex(ctx.key)
}
