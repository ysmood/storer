package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/postgres"
)

var db = postgres.New("")

func clean() {
	kit.E(db.Do(true, func(txn kvstore.Txn) error {
		return txn.Do(false, nil, func(k []byte) error {
			return txn.Delete(k)
		})
	}))
}

func TestBasic(t *testing.T) {
	clean()

	kit.E(db.Do(true, func(txn kvstore.Txn) error {
		err := txn.Set([]byte("key"), []byte("val"))
		kit.E(err)
		val, err := txn.Get([]byte("key"))
		kit.E(err)
		assert.Equal(t, []byte("val"), val)

		_, err = txn.Get([]byte("not-exists"))
		assert.Equal(t, err, kvstore.ErrKeyNotFound)

		return txn.Do(false, nil, func(key []byte) error {
			assert.Equal(t, []byte("key"), key)
			return nil
		})
	}))
}

func TestIteration(t *testing.T) {
	clean()

	list := [][]byte{}

	kit.E(db.Do(true, func(txn kvstore.Txn) error {
		for i := 0; i < 100; i++ {
			kit.E(txn.Set([]byte{byte(i)}, nil))
		}

		return txn.Do(false, nil, func(key []byte) error {
			list = append(list, key)
			return nil
		})
	}))

	assert.Len(t, list, 100)
	assert.Equal(t, []byte{99}, list[99])
}
