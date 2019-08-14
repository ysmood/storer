package storer_test

import (
	"encoding/binary"

	"github.com/ysmood/byframe"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/badger"
	"github.com/ysmood/storer/pkg/kvstore"
)

type User struct {
	Name  string
	Level int
}

var _ storer.Unique = &User{}
var _ storer.UniqueType = &User{}
var _ storer.Encoding = &User{}

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

func (u *User) Encode() ([]byte, error) {
	n := []byte(u.Name)

	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l, uint64(u.Level))
	return byframe.EncodeTuple(&n, &l), nil
}

func (u *User) Decode(data []byte) error {
	var name, level []byte
	kit.E(byframe.DecodeTuple(data, &name, &level))

	u.Level = int(binary.LittleEndian.Uint64(level))
	u.Name = string(name)
	return nil
}

type queue []interface{}

var empty = struct{}{}

func (s *queue) pop() interface{} {
	if s == nil || len(*s) == 0 {
		return empty
	}
	res := (*s)[0]
	*s = (*s)[1:]
	return res
}

func (s *queue) do(fn func() interface{}) interface{} {
	res := s.pop()
	if res == empty {
		return fn()
	}
	return res
}

type TestStore struct {
	badger *badger.Badger

	doQueue     queue
	doTxnQueue  queue
	getQueue    queue
	setQueue    queue
	deleteQueue queue
}

type TestTxn struct {
	store *TestStore
	txn   kvstore.Txn
}

var _ storer.Database = &TestStore{}
var _ kvstore.Txn = &TestTxn{}

func (s *TestStore) Close() error {
	return s.badger.Close()
}

func (s *TestStore) Do(update bool, fn kvstore.DoTxn) error {
	return s.badger.Do(update, func(txn kvstore.Txn) error {
		ret := s.doQueue.do(func() interface{} {
			return fn(&TestTxn{
				store: s,
				txn:   txn,
			})
		})
		err, _ := ret.(error)
		return err
	})
}

// Get when err is ErrKeyNotFound the key doesn't exist
func (txn *TestTxn) Get(key []byte) (value []byte, err error) {
	ret := txn.store.getQueue.do(func() interface{} {
		val, err := txn.txn.Get(key)
		return []interface{}{val, err}
	}).([]interface{})

	value, _ = ret[0].([]byte)
	err, _ = ret[1].(error)

	return
}

// Set set item with key and value
func (txn *TestTxn) Set(key, value []byte) error {
	ret := txn.store.setQueue.do(func() interface{} {
		return txn.txn.Set(key, value)
	})
	err, _ := ret.(error)
	return err
}

// Delete delete item via key
func (txn *TestTxn) Delete(key []byte) error {
	ret := txn.store.deleteQueue.do(func() interface{} {
		return txn.txn.Delete(key)
	})
	err, _ := ret.(error)
	return err
}

// Do key only iteration.
// The order of the iteration must be byte-wise lexicographical with the key.
// If fn returns an ErrStop the iteration will stop without error.
func (txn *TestTxn) Do(reverse bool, from []byte, fn kvstore.Iteratee) error {
	ret := txn.store.doTxnQueue.do(func() interface{} {
		return txn.txn.Do(reverse, from, fn)
	})
	err, _ := ret.(error)
	return err
}
