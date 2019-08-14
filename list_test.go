package storer_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestListCRUD(t *testing.T) {
	users := store.List(&User{})

	index := users.Index("name", func(u *User) interface{} {
		return u.Level
	})

	kit.E(store.Update(func(txn kvstore.Txn) error {
		usersTxn := users.Txn(txn)
		indexTxn := index.Txn(txn)

		var jack User

		id, err := usersTxn.Add(&User{"jack", 10})
		kit.E(err)

		kit.E(usersTxn.Get(id, &jack))
		assert.Equal(t, 10, jack.Level)

		kit.E(usersTxn.Set(id, &User{"jack", 20}))
		kit.E(usersTxn.Set(id, &User{"jack", 20}))

		kit.E(indexTxn.From(20).Find(&jack))

		assert.Equal(t, 20, jack.Level)

		kit.E(usersTxn.Del(id))

		assert.Equal(t, storer.ErrNotFound, indexTxn.From("jack").Find(&jack))

		return nil
	}))
}

func TestItemPtrErr(t *testing.T) {
	assert.PanicsWithValue(t, storer.ErrItemPtr, func() {
		_ = store.List(User{})
	})
}

func TestErrItemType(t *testing.T) {
	users := store.List(&User{})

	type MyUser struct{}
	_, err := users.Add(&MyUser{})
	assert.Equal(t, storer.ErrItemType, err)

	id, _ := users.Add(&User{})

	var user MyUser
	err = users.Get(id, &user)
	assert.Equal(t, storer.ErrItemType, err)
}

func TestHexErr(t *testing.T) {
	users := store.List(&User{})
	var user User
	assert.EqualError(t, users.Get(".", &user), "encoding/hex: invalid byte: U+002E '.'")
	assert.EqualError(t, users.Set(".", &user), "encoding/hex: invalid byte: U+002E '.'")
	assert.EqualError(t, users.Del("."), "encoding/hex: invalid byte: U+002E '.'")
}

type EncodeErr struct {
	err error
}

func (e *EncodeErr) Encode() ([]byte, error) {
	return nil, e.err
}

func (e *EncodeErr) Decode([]byte) error {
	return e.err
}

var _ storer.Encoding = &EncodeErr{}

func TestEncodeErr(t *testing.T) {
	err := errors.New("err")
	list := store.List(&EncodeErr{})
	_, e := list.Add(&EncodeErr{err})
	assert.Equal(t, err, e)
}

func TestListErrs(t *testing.T) {
	testErr := errors.New("err")

	db := &TestStore{badger: badger.New("")}
	store := storer.Store{
		DB: db,
	}
	users := store.List(&User{})

	id, _ := users.Add(&User{})

	db.getQueue = []interface{}{
		[]interface{}{[]byte{}, testErr},
	}

	err := users.Set(id, &User{})
	assert.Equal(t, testErr, err)

	db.setQueue = []interface{}{testErr}
	err = users.Set(id, &User{})
	assert.Equal(t, testErr, err)

	db.getQueue = []interface{}{
		[]interface{}{[]byte{}, testErr},
	}
	err = users.Del(id)
	assert.Equal(t, testErr, err)

	db.getQueue = []interface{}{
		[]interface{}{[]byte{}, testErr},
	}
	assert.PanicsWithValue(t, testErr, func() {
		_ = users.Index("", func(u *User) interface{} {
			return u.Level
		})
	})

	_ = users.UniqueIndex("", func(u *User) interface{} {
		return u.Level
	})
	db.doTxnQueue = []interface{}{testErr}
	_, err = users.Add(&User{})
	assert.Equal(t, testErr, err)
}
