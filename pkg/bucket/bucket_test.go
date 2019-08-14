package bucket_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestBasic(t *testing.T) {
	dir := "tmp/" + kit.RandString(10)

	db := badger.New(dir)
	_ = db.Do(true, func(txn kvstore.Txn) error {
		b, _ := bucket.New(txn, "test")
		assert.True(t, b.Valid([]byte{2}))

		b, _ = bucket.New(txn, "test")
		assert.True(t, b.Valid([]byte{2}))

		assert.Equal(t, 1, b.Len())

		return nil
	})

	// simulate to close the db file
	db.Close()

	// open the db file again
	_ = badger.New(dir).Do(true, func(txn kvstore.Txn) error {
		b, _ := bucket.New(txn, "test")
		assert.True(t, b.Valid([]byte{2}))

		b, _ = bucket.New(txn, "a")
		assert.True(t, b.Valid([]byte{3}))

		b, _ = bucket.New(txn, "")
		assert.True(t, b.Valid([]byte{4}))

		b, _ = bucket.New(txn, kit.RandString(10))
		_ = b.Set(txn, []byte("key"), []byte("value"))
		val, _ := b.Get(txn, []byte("key"))
		assert.Equal(t, []byte("value"), val)
		return nil
	})
}

type ErrTxn struct {
	bucket.Txn
	count int
	errs  []error
}

var errTxn = errors.New("err")

func (txn *ErrTxn) Get(key []byte) (value []byte, err error) {
	err = txn.errs[txn.count]
	txn.count++
	return nil, err
}

func (txn *ErrTxn) Set(key, value []byte) error {
	err := txn.errs[txn.count]
	txn.count++
	return err
}

func TestErr(t *testing.T) {
	txn := &ErrTxn{errs: []error{errTxn}}
	_, err := bucket.New(txn, "test")
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, errTxn}}
	_, err = bucket.New(txn, "test")
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, nil, errTxn}}
	_, err = bucket.New(txn, "test")
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, nil, nil, errTxn}}
	_, err = bucket.New(txn, "test")
	assert.Equal(t, errTxn, err)

	b := bucket.Bucket{}
	assert.False(t, b.Valid([]byte{129}))
}
