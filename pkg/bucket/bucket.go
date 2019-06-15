package bucket

// This bucket lib is dependency free, you can use whatever backend you want.

import (
	"bytes"

	"github.com/ysmood/byframe"
	"github.com/ysmood/storer/pkg/kvstore"
)

// the prefix for name map
var counterPrefix = byframe.EncodeHeader(0)
var nameMapPrefix = byframe.EncodeHeader(1)

// Txn the transaction interface
type Txn interface {
	// Get when err is ErrKeyNotFound the key doesn't exist
	Get(key []byte) (value []byte, err error)

	// Set ...
	Set(key, value []byte) error
}

// Bucket ...
type Bucket struct {
	prefix []byte
}

// New create an bucket via the name.
// For the same database file use the same name will always get the same prefix.
// The prefix size is dynamic, it begins with 1 byte.
// About the max number of buckets, the minimum is 2^(4*7 - 1) - 2,
// for 64bit CPU the number is 2^(8*7 - 1) - 2.
// LRU is used to take advantage of memory cache.
func New(txn Txn, name string) (*Bucket, error) {
	key := append(nameMapPrefix, name...)

	prefix, err := txn.Get(key)
	if err == nil {
		return &Bucket{prefix: prefix}, err
	} else if err != kvstore.ErrKeyNotFound {
		return nil, err
	}

	countData, err := txn.Get(counterPrefix)
	if err == kvstore.ErrKeyNotFound {
		countData = byframe.EncodeHeader(1)
	} else if err != nil {
		return nil, err
	}

	count, _, _ := byframe.DecodeHeader(countData)
	countData = byframe.EncodeHeader(count + 1)

	err = txn.Set(counterPrefix, countData)
	if err != nil {
		return nil, err
	}

	// count as prefix
	err = txn.Set(key, countData)
	if err != nil {
		return nil, err
	}
	return &Bucket{prefix: countData}, nil
}

// Set set key and value to the store with prefix
func (b *Bucket) Set(txn Txn, key, value []byte) error {
	return txn.Set(append(b.prefix, key...), value)
}

// Get get value by the key from the store with prefix
func (b *Bucket) Get(txn Txn, key []byte) ([]byte, error) {
	return txn.Get(append(b.prefix, key...))
}

// Prefix prefix key
func (b *Bucket) Prefix(key []byte) []byte {
	return append(b.prefix, key...)
}

// Len length of the prefix
func (b *Bucket) Len() int {
	return len(b.prefix)
}

// Valid check if the key is prefixed with the prefix.
func (b *Bucket) Valid(prefixedKey []byte) bool {
	_, l, sufficient := byframe.DecodeHeader(prefixedKey)
	if !sufficient {
		return false
	}
	return bytes.Equal(b.prefix, prefixedKey[:l])
}
