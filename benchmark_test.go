package storer_test

import (
	"testing"

	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/kvstore"

	"github.com/ysmood/kit"
)

func BenchmarkBadgerSet(b *testing.B) {
	store := storer.New("")

	data, err := storer.Encode(User{Name: "jack", Level: 1})
	kit.E(err)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(store.DB.Do(true, func(txn kvstore.Txn) error {
			return txn.Set(kit.RandBytes(12), data)
		}))
	}
}

func BenchmarkSet(b *testing.B) {
	store := storer.New("")

	users := store.List(&User{})

	user := &User{Name: "jack", Level: 1}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Add(user))
	}
}

func BenchmarkBadgerGet(b *testing.B) {
	store := storer.New("")

	data, err := storer.Encode(User{Name: "jack", Level: 1})
	kit.E(err)

	var id []byte
	for i := 0; i < 10000; i++ {
		var err error
		kit.E(store.DB.Do(true, func(txn kvstore.Txn) error {
			id = kit.RandBytes(12)
			return txn.Set(id, data)
		}))
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(store.DB.Do(false, func(txn kvstore.Txn) error {
			_, err := txn.Get(id)
			return err
		}))
	}
}

func BenchmarkGet(b *testing.B) {
	store := storer.New("")

	users := store.List(&User{})

	user := &User{Name: "jack", Level: 1}

	var id string
	for i := 0; i < 10000; i++ {
		var err error
		id, err = users.Add(user)
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Get(id, user))
	}
}

func BenchmarkGetByIndex(b *testing.B) {
	users := store.List(&User{})
	index := users.Index("level", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	for i := 0; i < 10000; i++ {
		kit.E(users.Add(&User{
			Name:  "jack",
			Level: i,
		}))
	}

	var item User

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(index.From(9999).Find(&item))
	}
}

func BenchmarkFilter(b *testing.B) {
	users := store.List(&User{})
	index := users.Index("level", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	for i := 0; i < 10000; i++ {
		kit.E(users.Add(&User{
			Name:  "jack",
			Level: i,
		}))
	}

	var items []User

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(index.From(9990).Filter(&items, func(_ *storer.IterCtx) (bool, bool) {
			return true, true
		}))
	}
}
