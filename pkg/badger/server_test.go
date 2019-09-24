package badger_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestServerCRUD(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	kit.E(err)

	db := badger.New("")

	go func() {
		kit.E(badger.Serve(db, l))
	}()

	client := badger.NewClient(l.Addr().String())

	kit.E(client.Do(true, func(txn kvstore.Txn) error {
		_ = txn.Set([]byte("a"), []byte("b"))

		val, err := txn.Get([]byte("a"))
		kit.E(err)

		assert.Equal(t, []byte("b"), val)

		kit.E(txn.Delete([]byte("a")))

		_, err = txn.Get([]byte("a"))
		assert.Equal(t, kvstore.ErrKeyNotFound, err)

		return nil
	}))
}

func TestServerIteration(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	kit.E(err)

	db := badger.New("")

	go func() {
		kit.E(badger.Serve(db, l))
	}()

	client := badger.NewClient(l.Addr().String())

	keys := []byte{}
	values := []byte{}

	kit.E(client.Do(true, func(txn kvstore.Txn) error {
		_ = txn.Set([]byte("1"), []byte("a"))
		_ = txn.Set([]byte("2"), []byte("b"))
		_ = txn.Set([]byte("3"), []byte("c"))

		return txn.Do(false, nil, func(key []byte) error {
			keys = append(keys, key[0])
			val, err := txn.Get(key)
			if err != nil {
				return err
			}
			values = append(values, val[0])
			return nil
		})
	}))

	assert.Equal(t, []byte{'1', '2', '3'}, keys)
	assert.Equal(t, []byte{'a', 'b', 'c'}, values)
}
