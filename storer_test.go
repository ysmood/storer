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
