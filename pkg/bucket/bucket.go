package bucket

import (
	"bytes"
	"errors"

	"github.com/ysmood/byframe"
	"github.com/ysmood/storer/pkg/kvstore"
)

// the prefix for name map
var nameMapPrefix = byframe.EncodeHeader(0)

// ErrKeyNotFound ...
var ErrKeyNotFound = kvstore.ErrKeyNotFound

// ErrEmptyName ...
var ErrEmptyName = errors.New("bucket: name cannot be empty")

// Txn ...
type Txn = kvstore.Txn

// Bucket ...
type Bucket struct {
	prefix []byte
}

// New create an bucket via the name.
// For the same database file use the same name will always get the same bucket.
// The prefix size is dynamic, it begins with 1 byte.
// About the max number of buckets is usually based on the CPU, for 32bit CPU it's 2^(4*7 - 1) - 1,
// for 64bit CPU the number is 2^(8*7 - 1) - 1. It's way enough for common usage.
func New(txn Txn, name []byte) (*Bucket, error) {
	if len(name) == 0 {
		return nil, ErrEmptyName
	}

	key := append(nameMapPrefix, name...)

	prefix, err := txn.Get(key)
	if err == nil {
		return &Bucket{prefix}, nil
	} else if err != ErrKeyNotFound {
		return nil, err
	}

	countData, err := txn.Get(nameMapPrefix)

	if err == ErrKeyNotFound {
		countData = nameMapPrefix
	} else if err != nil {
		return nil, err
	}

	count, _, _ := byframe.DecodeHeader(countData)
	countData = byframe.EncodeHeader(count + 1)

	err = txn.Set(nameMapPrefix, countData)
	if err != nil {
		return nil, err
	}

	// count as prefix
	err = txn.Set(key, countData)
	if err != nil {
		return nil, err
	}
	return &Bucket{countData}, nil
}

// Delete delete bucket from the db
func Delete(txn Txn, name string) error {
	return txn.Delete(append(nameMapPrefix, name...))
}

// Set set key and value to the store with prefix
func (b *Bucket) Set(txn Txn, key, value []byte) error {
	return txn.Set(b.Prefix(key), value)
}

// Get get value by the key from the store with prefix
func (b *Bucket) Get(txn Txn, key []byte) ([]byte, error) {
	return txn.Get(b.Prefix(key))
}

// Delete delete value by the key with prefix
func (b *Bucket) Delete(txn Txn, key []byte) error {
	return txn.Delete(b.Prefix(key))
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

// Empty remove everything in the bucket.
func (b *Bucket) Empty(txn Txn) error {
	return txn.Do(false, b.prefix, func(key []byte) error {
		if b.Valid(key) {
			return txn.Delete(key)
		}
		return kvstore.ErrStop
	})
}
