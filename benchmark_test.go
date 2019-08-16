package storer_test

import (
	"testing"

	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/typee"

	"github.com/ysmood/kit"
)

const rounds = 10000

func BenchmarkBadgerSet(b *testing.B) {
	db := badger.New("")

	data, err := typee.Encode(&User{Name: "jack", Level: 1}, nil)
	kit.E(err)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(db.Do(true, func(txn kvstore.Txn) error {
			return txn.Set(kit.RandBytes(12), data)
		}))
	}
}

func BenchmarkSet(b *testing.B) {
	store := storer.New("")

	users := store.ListWithName(kit.RandString(10), &User{})

	user := &User{Name: "jack", Level: 1}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Add(user))
	}
}

func BenchmarkSetWithIndex(b *testing.B) {
	store := storer.New("")

	users := store.ListWithName(kit.RandString(10), &User{})
	_ = users.Index("level", func(u *User) interface{} {
		return u.Level
	})

	user := &User{Name: "jack", Level: 1}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Add(user))
	}
}
func BenchmarkBadgerGet(b *testing.B) {
	db := badger.New("")

	data, err := typee.Encode(&User{Name: "jack", Level: 1}, nil)
	kit.E(err)

	var id []byte
	for i := 0; i < rounds; i++ {
		var err error
		kit.E(db.Do(true, func(txn kvstore.Txn) error {
			id = kit.RandBytes(12)
			return txn.Set(id, data)
		}))
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(db.Do(false, func(txn kvstore.Txn) error {
			_, err := txn.Get(id)
			return err
		}))
	}
}

func BenchmarkGet(b *testing.B) {
	store := storer.New("")

	users := store.ListWithName(kit.RandString(10), &User{})

	user := &User{Name: "jack", Level: 1}

	var id string
	for i := 0; i < rounds; i++ {
		var err error
		id, err = users.Add(user)
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(users.Get(id, user))
	}
}

func BenchmarkBadgerPrefixGet(b *testing.B) {
	db := badger.New("")

	data, err := typee.Encode(&User{Name: "jack", Level: 1}, nil)
	kit.E(err)

	var id []byte
	for i := 0; i < rounds; i++ {
		var err error
		kit.E(db.Do(true, func(txn kvstore.Txn) error {
			id = kit.RandBytes(12)
			return txn.Set(id, data)
		}))
		kit.E(err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(db.Do(false, func(txn kvstore.Txn) error {
			return txn.Do(false, id, func(key []byte) error {
				_, _ = txn.Get(id)
				return storer.ErrStop
			})
		}))
	}
}

func BenchmarkGetByIndex(b *testing.B) {
	users := store.ListWithName(kit.RandString(10), &User{})
	index := users.Index("level", func(u *User) interface{} {
		return u.Level
	})

	for i := 0; i < rounds; i++ {
		kit.E(users.Add(&User{
			Name:  "jack",
			Level: i,
		}))
	}

	var item User

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(index.From(rounds - 1).Find(&item))
	}
}

func BenchmarkFilter(b *testing.B) {
	users := store.ListWithName(kit.RandString(10), &User{})
	index := users.Index("level", func(u *User) interface{} {
		return u.Level
	})

	for i := 0; i < rounds; i++ {
		kit.E(users.Add(&User{
			Name:  "jack",
			Level: i,
		}))
	}

	var items []User

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		kit.E(index.From(9990).Filter(&items, func(_ *User) interface{} {
			return true
		}))
	}
}
