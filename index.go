package storer

import (
	"bytes"

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

	return txn.Set(index.id(i, itemID), nil)
}

func (index *Index) update(txn kvstore.Txn, itemID []byte, old, new interface{}) error {
	oldIndex, err := index.genIndex(&GenCtx{Item: old, Txn: txn, Action: IndexUpdate})
	if err != nil {
		return err
	}
	newIndex, err := index.genIndex(&GenCtx{Item: new, Txn: txn, Action: IndexUpdate})
	if err != nil {
		return err
	}

	if bytes.Equal(oldIndex, newIndex) {
		return nil
	}

	err = txn.Delete(index.id(oldIndex, itemID))
	if err != nil {
		return err
	}

	return txn.Set(index.id(newIndex, itemID), nil)
}

func (index *Index) del(txn kvstore.Txn, itemID []byte, item interface{}) error {
	i, err := index.genIndex(&GenCtx{Item: item, Txn: txn, Action: IndexDelete})
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
	prefix := ctx.txnCtx.index.bucket.Prefix(byframe.Encode(ctx.from))

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
