package badger_test

import (
	"errors"
	"testing"

	originBadger "github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestBasic(t *testing.T) {
	db := badger.New("")

	_ = db.Do(true, func(txn kvstore.Txn) error {
		k, v := []byte("k"), []byte("v")

		_ = txn.Set(k, v)

		gv, _ := txn.Get(k)
		assert.Equal(t, v, gv)

		_ = txn.Do(false, nil, func(gk []byte) error {
			assert.Equal(t, k, gk)

			gv, _ := txn.Get(gk)
			assert.Equal(t, v, gv)

			return kvstore.ErrStop
		})

		_ = txn.Delete(k)

		_, err := txn.Get(k)
		assert.Equal(t, kvstore.ErrKeyNotFound, err)

		return nil
	})

	db.Close()
}

func TestIteration(t *testing.T) {
	db := badger.New("")

	a := []byte("a")
	b := []byte("b")
	c := []byte("c")

	_ = db.Do(true, func(txn kvstore.Txn) error {
		_ = txn.Set(a, c)
		_ = txn.Set(b, a)
		_ = txn.Set(c, b)

		return nil
	})

	_ = db.Do(false, func(txn kvstore.Txn) error {
		keys := [][]byte{}
		values := [][]byte{}

		_ = txn.Do(false, b, func(k []byte) error {
			keys = append(keys, k)

			v, _ := txn.Get(k)
			values = append(values, v)

			return nil
		})

		assert.Equal(t, [][]byte{b, c}, keys)
		assert.Equal(t, [][]byte{a, b}, values)

		return nil
	})
}

func TestErr(t *testing.T) {
	db := badger.New("")

	testErr := errors.New("err")

	err := db.Do(false, func(txn kvstore.Txn) error {
		return testErr
	})
	assert.Equal(t, testErr, err)

	err = db.Do(true, func(txn kvstore.Txn) error {
		_ = txn.Set([]byte("k"), nil)
		return txn.Do(false, nil, func(_ []byte) error {
			return testErr
		})
	})
	assert.Equal(t, testErr, err)

	err = db.Do(true, func(txn kvstore.Txn) error {
		_, err := txn.Get(nil)
		return err
	})
	assert.Equal(t, originBadger.ErrEmptyKey, err)
}
