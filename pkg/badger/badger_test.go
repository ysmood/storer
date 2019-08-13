package badger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

func do(dir string, fn func(db kvstore.Store)) {
	kit.E(kit.Mkdir(dir, nil))

	db := badger.New(dir)

	fn(db)

	db.Close()
}

func TestBasic(t *testing.T) {
	do("tmp/"+kit.RandString(10), func(db kvstore.Store) {
		_ = db.Do(true, func(txn kvstore.Txn) error {
			k, v := []byte("k"), []byte("v")

			_ = txn.Set(k, v)

			gv, _ := txn.Get(k)
			assert.Equal(t, v, gv)

			_ = txn.Do(false, nil, func(ctx kvstore.IterCtx) error {
				gk := ctx.Key()
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
	})
}

func TestSeek(t *testing.T) {
	do("tmp/"+kit.RandString(10), func(db kvstore.Store) {
		_ = db.Do(true, func(txn kvstore.Txn) error {
			a := []byte("a")
			b := []byte("b")
			c := []byte("c")

			_ = txn.Set(a, c)
			_ = txn.Set(b, a)
			_ = txn.Set(c, b)

			keys := [][]byte{}
			values := [][]byte{}

			first := true
			_ = txn.Do(false, nil, func(ctx kvstore.IterCtx) error {
				if first {
					ctx.Seek(b)
					first = false
				}

				keys = append(keys, ctx.Key())

				v, _ := txn.Get(ctx.Key())
				values = append(values, v)

				return nil
			})

			assert.Equal(t, [][]byte{b, c}, keys)
			assert.Equal(t, [][]byte{a, b}, values)

			return nil
		})
	})
}
