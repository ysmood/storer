package bucket_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/bucket"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestBasic(t *testing.T) {
	dir := "tmp/" + kit.RandString(10)
	kit.E(kit.Mkdir(dir, nil))

	db := badger.New(dir)
	_ = db.Do(true, func(txn kvstore.Txn) error {
		b, _ := bucket.New(txn, "test")
		assert.True(t, b.Valid([]byte{2}))

		b, _ = bucket.New(txn, "test")
		assert.True(t, b.Valid([]byte{2}))

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
