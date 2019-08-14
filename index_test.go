package storer_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ysmood/kit"
	"github.com/ysmood/storer"
)

func TestNotFound(t *testing.T) {
	users := store.List(&User{})
	index := users.Index("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})

	_, _ = users.Add(&User{"jack", 10})

	items := []User{}
	err := index.From("jaca").Find(&items)
	assert.Equal(t, storer.ErrNotFound, err)

	err = index.From("jac").Find(&items)
	assert.Equal(t, storer.ErrNotFound, err)
}

func TestLongIndex(t *testing.T) {
	longStr := kit.RandString(1000)

	users := store.List(&User{})
	index := users.Index("name", func(u *User) interface{} {
		return longStr + u.Name
	})

	_, _ = users.Add(&User{"jack", 10})

	var user User
	kit.E(index.From(longStr + "jack").Find(&user))
	assert.Equal(t, "jack", user.Name)
}

func TestReverse(t *testing.T) {
	users := store.List(&User{})
	index := users.Index("name", func(u *User) interface{} {
		return u.Level
	})

	_, _ = users.Add(&User{"jack", 1})
	_, _ = users.Add(&User{"jack", 2})

	_ = store.View(func(txn storer.Txn) error {
		items := []User{}
		kit.E(index.Txn(txn).From(3).Reverse().Filter(&items, func(_ *User) interface{} {
			return true
		}))
		assert.Equal(t, 2, items[0].Level)
		assert.Equal(t, 1, items[1].Level)
		return nil
	})
}

func TestFindAll(t *testing.T) {
	users := store.List(&User{})
	index := users.Index("name", func(u *User) interface{} {
		return u.Name
	})

	_, _ = users.Add(&User{"jack", 10})
	_, _ = users.Add(&User{"jack", 20})

	items := []User{}
	err := index.From("jack").Find(&items)
	kit.E(err)

	assert.Len(t, items, 2)
	assert.Equal(t, 30, items[0].Level+items[1].Level)
}

func TestEach(t *testing.T) {
	users := store.List(&User{})
	index := users.Index("age", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	_, _ = users.Add(&User{"jack", 10})
	_, _ = users.Add(&User{"jack", 20})

	levels := []int{}
	_ = index.From(0).Each(func(ctx *storer.IterCtx) error {
		var user User
		_ = ctx.Item(&user)
		levels = append(levels, user.Level)
		return nil
	})
	assert.Equal(t, []int{10, 20}, levels)
}

func TestFilter(t *testing.T) {
	users := store.List(&User{})
	index := users.Index("age", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	// insert random level into the list, the range is [-10, 10)
	for _, i := range rand.Perm(20) {
		_, _ = users.Add(&User{"a", i - 10})
	}

	items := []User{}

	// get range [-7, 2)
	kit.E(index.From(-7).Filter(&items, func(u *User) interface{} {
		if u.Level < 2 {
			return true
		}
		return storer.ErrStop
	}))

	assert.Len(t, items, 9)
	assert.Equal(t, -7, items[0].Level)
	assert.Equal(t, -5, items[2].Level)
	assert.Equal(t, 1, items[8].Level)
}

func TestUniqueIndex(t *testing.T) {
	users := store.List(&User{})
	_ = users.UniqueIndex("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})

	id, _ := users.Add(&User{"jack", 10})
	assert.Nil(t, users.Set(id, &User{"jack", 20}))
	_, err := users.Add(&User{"jack", 20})

	assert.Equal(t, storer.ErrUniqueIndex, err)

	_ = users.UniqueIndex("return err", func(ctx *storer.GenCtx) interface{} {
		return errors.New("err")
	})
	_, err = users.Add(&User{"ann", 10})
	assert.EqualError(t, err, "err")
}

func TestIndexExistsErr(t *testing.T) {
	users := store.List(&User{})
	_ = users.Index("name", func(_ *User) interface{} {
		return nil
	})
	assert.PanicsWithValue(t, storer.ErrIndexExists, func() {
		_ = users.Index("name", func(_ *User) interface{} {
			return nil
		})
	})
}
