package storer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer"
)

var store *storer.Store

func TestMain(m *testing.M) {
	dir := "tmp/" + kit.RandString(10)

	store = storer.New(dir)

	os.Exit(m.Run())
}

func TestClose(t *testing.T) {
	store := storer.New("")
	assert.Nil(t, store.Close())
}

func TestValue(t *testing.T) {
	type myval int
	var v myval = 1
	val := store.Value(kit.RandString(10), &v)

	_ = val.Get(&v)

	v++
	_ = val.Set(&v)

	v = 0
	_ = val.Get(&v)
	assert.Equal(t, myval(2), v)

	_ = store.Update(func(txn storer.Txn) error {
		v = 10
		_ = val.Txn(txn).Set(&v)

		v = 0
		_ = val.Txn(txn).Get(&v)
		assert.Equal(t, myval(10), v)

		return nil
	})
}
