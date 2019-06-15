package storer_test

import (
	"encoding/binary"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	kit "github.com/ysmood/gokit"
	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/kvstore"
)

type User struct {
	Name  string
	Level int
}

var UserIDCounter = uint64(0)

func (u *User) ID() []byte {
	b := make([]byte, 8)
	n := binary.PutUvarint(b, UserIDCounter)
	UserIDCounter++
	return b[:n]
}

func (u *User) TypeID() string {
	return kit.RandString(10)
}

var store *storer.Store

func TestMain(m *testing.M) {
	dir := "tmp/" + kit.RandString(10)

	kit.E(kit.Mkdir(dir, nil))

	store = storer.New(dir)
	defer store.DB.Close()

	os.Exit(m.Run())
}

func TestMap(t *testing.T) {
	users, err := store.Map(&User{})
	kit.E(err)

	kit.E(store.Update(func(txn kvstore.Txn) error {
		usersTxn := users.Txn(txn)

		var jack User

		key := "key"

		kit.E(usersTxn.Set(key, &User{"jack", 10}))

		kit.E(usersTxn.Get(key, &jack))

		assert.Equal(t, 10, jack.Level)

		kit.E(usersTxn.Del(key))

		assert.Equal(t, kvstore.ErrKeyNotFound, usersTxn.Get(key, &jack))

		return nil
	}))
}

func TestListCRUD(t *testing.T) {
	users, err := store.List(&User{})
	kit.E(err)

	index, err := users.Index("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})
	kit.E(err)

	kit.E(store.Update(func(txn kvstore.Txn) error {
		usersTxn := users.Txn(txn)
		indexTxn := index.Txn(txn)

		var jack User

		id, err := usersTxn.Add(&User{"jack", 10})
		kit.E(err)

		kit.E(usersTxn.Get(id, &jack))
		assert.Equal(t, 10, jack.Level)

		kit.E(usersTxn.Set(id, &User{"jack", 20}))

		kit.E(indexTxn.For("jack").Find(&jack))

		assert.Equal(t, 20, jack.Level)

		kit.E(usersTxn.Del(id))

		assert.Equal(t, storer.ErrNotFound, indexTxn.For("jack").Find(&jack))

		return nil
	}))
}

func TestItemPtrErr(t *testing.T) {
	_, err := store.List(User{})

	assert.Equal(t, storer.ErrItemPtr, err)
}

func TestNotFound(t *testing.T) {
	users, _ := store.List(&User{})
	index, _ := users.Index("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})

	_, _ = users.Add(&User{"jack", 10})

	items := []User{}
	err := index.For("jaca").Find(&items)
	assert.Equal(t, storer.ErrNotFound, err)

	err = index.For("jac").Find(&items)
	assert.Equal(t, storer.ErrNotFound, err)
}

func TestLongIndex(t *testing.T) {
	longStr := kit.RandString(1000)

	users, _ := store.List(&User{})
	index, _ := users.Index("name", func(ctx *storer.GenCtx) interface{} {
		return longStr + ctx.Item.(*User).Name
	})

	_, _ = users.Add(&User{"jack", 10})

	var user User
	kit.E(index.For(longStr + "jack").Find(&user))
	assert.Equal(t, "jack", user.Name)
}

func TestFindAll(t *testing.T) {
	users, _ := store.List(&User{})
	index, _ := users.Index("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})

	_, _ = users.Add(&User{"jack", 10})
	_, _ = users.Add(&User{"jack", 20})

	items := []User{}
	err := index.For("jack").Find(&items)
	kit.E(err)

	assert.Len(t, items, 2)
	assert.Equal(t, 30, items[0].Level+items[1].Level)
}

func TestRange(t *testing.T) {
	users, _ := store.List(&User{})
	index, _ := users.Index("age", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	// insert random level into the list, the range is [-10, 10)
	for _, i := range rand.Perm(20) {
		_, _ = users.Add(&User{"a", i - 10})
	}

	items := []User{}

	// get range [-7, 2)
	kit.E(index.For(-7).Range(&items, func(ctx *storer.IterCtx) bool {
		return ctx.Compare(2) < 0
	}))

	assert.Len(t, items, 9)
	assert.Equal(t, -7, items[0].Level)
	assert.Equal(t, -5, items[2].Level)
	assert.Equal(t, 1, items[8].Level)
}

func TestUniqueIndex(t *testing.T) {
	users, _ := store.List(&User{})
	_, _ = users.UniqueIndex("name", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Name
	})

	_, _ = users.Add(&User{"jack", 10})
	_, err := users.Add(&User{"jack", 20})

	assert.Equal(t, storer.ErrUniqueIndex, err)
}
