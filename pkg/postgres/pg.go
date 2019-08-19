package postgres

import (
	"database/sql"
	"fmt"

	"github.com/ysmood/kit/pkg/utils"
	"github.com/ysmood/storer/pkg/kvstore"

	_ "github.com/lib/pq" // pg driver
)

// PG adapter
type PG struct {
	db *sql.DB

	// PrefetchSize by default each time 50 keys will be fetched
	PrefetchSize int
}

var _ kvstore.Store = &PG{}

// New a helper to create an adapter instance.
// If the connStr is empty a random database will be created.
func New(connStr string) *PG {
	var dbName string
	if connStr == "" {
		dbName = utils.RandString(10)
		connStr = "user=postgres sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	utils.E(err)

	if dbName != "" {
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, dbName))
		utils.E(err)
		utils.E(db.Close())
		db, err = sql.Open("postgres", fmt.Sprintf(`user=postgres sslmode=disable dbname=%s`, dbName))
		utils.E(err)
	}

	_, err = db.Query(`
		CREATE TABLE IF NOT EXISTS store (
			key bytea,
			val bytea
		);
		CREATE INDEX IF NOT EXISTS idx_key ON store(key);
	`)
	utils.E(err)

	return NewByDB(db)
}

// NewByDB ...
func NewByDB(db *sql.DB) *PG {
	return &PG{
		db:           db,
		PrefetchSize: 50,
	}
}

// Do ...
func (pg *PG) Do(update bool, fn kvstore.DoTxn) error {
	txn, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = txn.Rollback()
	}()

	err = fn(&Txn{
		db:  pg,
		txn: txn,
	})
	if err != nil {
		return err
	}

	if update {
		return txn.Commit()
	}
	return err
}

// Close ...
func (pg *PG) Close() error {
	return pg.db.Close()
}

// Txn ...
type Txn struct {
	db  *PG
	txn *sql.Tx
}

var _ kvstore.Txn = &Txn{}

// Get ...
func (t *Txn) Get(key []byte) ([]byte, error) {
	var val []byte
	err := t.txn.QueryRow(
		"SELECT val FROM store WHERE key = $1",
		key,
	).Scan(&val)

	if err == sql.ErrNoRows {
		return nil, kvstore.ErrKeyNotFound
	}

	return val, err
}

// Set ...
func (t *Txn) Set(key, value []byte) error {
	_, err := t.txn.Exec(
		`INSERT INTO store (key, val) VALUES ($1, $2)`,
		key, value,
	)
	return err
}

// Delete ...
func (t *Txn) Delete(key []byte) error {
	_, err := t.txn.Exec(
		`DELETE FROM store WHERE key = $1`,
		key,
	)
	return err
}

func (t *Txn) getKeys(reverse bool, from []byte, keys *[][]byte) error {
	var sql string
	if reverse {
		sql = `SELECT key FROM store WHERE key < $1 ORDER BY key DESC LIMIT $2`
	} else {
		sql = `SELECT key FROM store WHERE key > $1 ORDER BY key LIMIT $2`
	}

	rows, err := t.txn.Query(sql, from, t.db.PrefetchSize)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
	}()

	for rows.Next() {
		var key []byte
		err = rows.Scan(&key)
		if err != nil {
			break
		}
		*keys = append(*keys, key)
	}

	return err
}

// Do ...
func (t *Txn) Do(reverse bool, from []byte, fn kvstore.Iteratee) error {
	_, err := t.Get(from)
	if err == nil {
		err = fn(from)
		if err == kvstore.ErrStop {
			return nil
		}
		if err != nil {
			return err
		}
	} else if err != kvstore.ErrKeyNotFound {
		return err
	}

	for {
		var keys [][]byte
		err := t.getKeys(reverse, from, &keys)
		if err != nil {
			return err
		}

		for _, from = range keys {
			err = fn(from)
			if err == kvstore.ErrStop {
				return nil
			}
			if err != nil {
				break
			}
		}

		if len(keys) < t.db.PrefetchSize {
			break
		}
	}
	return nil
}
