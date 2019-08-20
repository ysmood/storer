package kdb

import (
	"errors"
	"io"
)

/*

Do one thing and do it well.

Key-only db may sound crazy, but it's a good idea to abstract away the side effects:

- snapshot transactions between machines
- store data in a sorted manner among machines
- traverse through sorted data from different machines

So that we can build high level functionalities based on of it with side effect free code.
Such as how to build key-value db on top of key-only db, introduce a side effect free
key-value encoding algorithm will solve it.

*/

// Store the persistent store for value
type Store interface {
	// Do do a new ACID transaction. Create a snapshot of the db.
	// Because we use snapshot for each txn, no need to specify readonly operation.
	Do(fn DoTxn) error
}

// DoTxn If it returns error the transaction will drop the snapshot.
// After it ends the snapshot should be merged into the db
type DoTxn func(txn Txn) error

// Txn the transaction interface
type Txn interface {
	// Add add data, large data will be distributed into small pieces,
	// so there should be no size limit for the data to store.
	Add(data io.Reader) error

	// Do Iterate the data by the byte-wise lexicographical order.
	Do(ascend bool, prefix io.Reader, fn Iteratee) error
}

// Iteratee If it returns ErrDelete the data will be removed from the db.
// If it returns ErrStop the walk will stop.
type Iteratee func(data io.ReadWriteSeeker) error

// ErrDelete ...
var ErrDelete = errors.New("kdb: delete data")

// ErrStop ...
var ErrStop = errors.New("kdb: stop walk")
