package storer_test

import (
	"testing"

	kit "github.com/ysmood/gokit"
	"github.com/ysmood/storer"
)

func BenchmarkSet(b *testing.B) {
	users, _ := store.List(&User{})

	user := &User{Name: "jack", Level: 1}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Add(user))
	}
}

func BenchmarkGet(b *testing.B) {
	users, _ := store.List(&User{})

	user := &User{Name: "jack", Level: 1}

	var id string
	for i := 0; i < 1000; i++ {
		var err error
		id, err = users.Add(user)
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Get(id, user))
	}
}

func BenchmarkRange(b *testing.B) {
	users, _ := store.List(&User{})
	index, _ := users.Index("level", func(ctx *storer.GenCtx) interface{} {
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
		kit.E(index.For(9990).Range(&items, func(_ *storer.IterCtx) bool {
			return true
		}))
	}
}
