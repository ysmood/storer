package storer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer/pkg/kvstore"
)

func TestMapTxn(t *testing.T) {
	users := store.Map(&User{})

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

func TestMapOps(t *testing.T) {
	users := store.Map(&User{})

	var jack User

	key := "key"

	kit.E(users.Set(key, &User{"jack", 10}))

	kit.E(users.Get(key, &jack))

	assert.Equal(t, 10, jack.Level)

	kit.E(users.Del(key))

	assert.Equal(t, kvstore.ErrKeyNotFound, users.Get(key, &jack))
}
