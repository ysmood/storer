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
		b, _ := bucket.New(txn, []byte("test"))
		assert.True(t, b.Valid([]byte{1}))

		b, _ = bucket.New(txn, []byte("test"))
		assert.True(t, b.Valid([]byte{1}))

		assert.Equal(t, 1, b.Len())

		return nil
	})

	// simulate to close the db file
	db.Close()

	// open the db file again
	_ = badger.New(dir).Do(true, func(txn kvstore.Txn) error {
		b, _ := bucket.New(txn, []byte("test"))
		assert.True(t, b.Valid([]byte{1}))

		b, _ = bucket.New(txn, []byte("a"))
		assert.True(t, b.Valid([]byte{2}))

		_, err := bucket.New(txn, []byte(""))
		assert.Equal(t, bucket.ErrEmptyName, err)

		b, _ = bucket.New(txn, kit.RandBytes(10))
		_ = b.Set(txn, []byte("key"), []byte("value"))
		val, _ := b.Get(txn, []byte("key"))
		assert.Equal(t, []byte("value"), val)

		_ = b.Delete(txn, []byte("key"))
		_, err = b.Get(txn, []byte("key"))
		assert.Equal(t, bucket.ErrKeyNotFound, err)
		return nil
	})
}

func TestDrop(t *testing.T) {
	dir := "tmp/" + kit.RandString(10)

	db := badger.New(dir)
	_ = db.Do(true, func(txn kvstore.Txn) error {
		a, _ := bucket.New(txn, []byte("a"))
		b, _ := bucket.New(txn, []byte("b"))

		_ = a.Set(txn, []byte("k"), []byte("a"))
		_ = b.Set(txn, []byte("k"), []byte("b"))

		_ = a.Empty(txn)
		_ = bucket.Delete(txn, "a")

		res := [][]byte{}
		_ = txn.Do(false, []byte{}, func(k []byte) error {
			kk := make([]byte, len(k))
			copy(kk, k)
			res = append(res, kk)
			return nil
		})
		assert.Len(t, res, 3)

		v, _ := b.Get(txn, []byte("k"))
		assert.Equal(t, []byte("b"), v)
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
	_, err := bucket.New(txn, []byte("test1"))
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, errTxn}}
	_, err = bucket.New(txn, []byte("test2"))
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, nil, errTxn}}
	_, err = bucket.New(txn, []byte("test3"))
	assert.Equal(t, errTxn, err)

	txn = &ErrTxn{errs: []error{bucket.ErrKeyNotFound, nil, nil, errTxn}}
	_, err = bucket.New(txn, []byte("test4"))
	assert.Equal(t, errTxn, err)

	b := bucket.Bucket{}
	assert.False(t, b.Valid([]byte{129}))
}
