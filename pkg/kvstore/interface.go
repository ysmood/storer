package kvstore

import "errors"

// The minimum interface to create an efficient indexable database with a key-value store.
// When any reduction of the interface design will dramatically impact the performance we can say the
// design is stable.
//
// The reason to use callback style is because some database may use resources that need to be cleaned up
// after certain operations, such as the lock to avoid race condition. Callback can abstract them away.
// The bottleneck is alway the underlaying syscalls, not the creation of callback functions.

// Store the persistent store for value
type Store interface {
	// Do do a new transaction.
	Do(update bool, fn DoTxn) error
}

// DoTxn If it returns error the transaction will discard the changes
type DoTxn func(txn Txn) error

// ErrKeyNotFound ...
var ErrKeyNotFound = errors.New("[storer.kvstore] key doesn't exist")

// Txn the transaction interface
type Txn interface {
	// Get an item. when err is ErrKeyNotFound the key doesn't exist
	Get(key []byte) (value []byte, err error)

	// Set set item with key and value
	Set(key, value []byte) error

	// Delete delete item via key
	Delete(key []byte) error

	// Do key only iteration.
	// The order of the iteration must be byte-wise lexicographical with the key.
	// If fn returns an ErrStop the iteration will stop without error.
	Do(reverse bool, from []byte, fn Iteratee) error
}

// ErrStop ...
var ErrStop = errors.New("[storer.kvstore] stop iteration")

// Iteratee ...
type Iteratee func(key []byte) error
