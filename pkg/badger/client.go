package badger

import (
	"errors"
	"net"

	"github.com/ysmood/byframe"
	"github.com/ysmood/storer/pkg/kvstore"
)

// Client ...
type Client struct {
	host string
}

var _ kvstore.Store = &Client{}

// NewClient ...
func NewClient(host string) *Client {
	return &Client{host: host}
}

func boolToBytes(b bool) *[]byte {
	if b {
		return &trueBytes
	}
	return &falseBytes
}

func bytesToBool(b []byte) bool {
	return b[0] == 1
}

// Do ...
func (c *Client) Do(update bool, fn kvstore.DoTxn) error {
	conn, err := net.Dial("tcp", c.host)
	if err != nil {
		return err
	}
	defer conn.Close()

	txn := &ClientTxn{
		conn:    conn,
		scanner: byframe.NewScanner(conn),
	}

	_, err = txn.request(actionBegin, boolToBytes(update))
	if err != nil {
		return err
	}

	err = fn(txn)
	if err != nil {
		return err
	}

	_, err = txn.request(actionEnd)
	return err
}

// ClientTxn ...
type ClientTxn struct {
	conn    net.Conn
	scanner *byframe.Scanner
}

var _ kvstore.Txn = &ClientTxn{}

func (t *ClientTxn) read() ([]byte, error) {
	if !t.scanner.Scan() {
		err := t.scanner.Err()
		if err != nil {
			return nil, err
		}
	}
	var resData, errData []byte
	err := byframe.DecodeTuple(t.scanner.Frame(), &resData, &errData)
	if err != nil {
		return nil, err
	}

	if len(errData) != 0 {
		err = errors.New(string(errData))
	}
	return resData, err
}

func (t *ClientTxn) request(action actionEnum, args ...*[]byte) ([]byte, error) {
	if len(args) == 0 {
		args = append([]*[]byte{action.bytes()}, &emptyBytes)
	} else {
		args = append([]*[]byte{action.bytes()}, args...)
	}
	_, err := t.conn.Write(byframe.Encode(byframe.EncodeTuple(args...)))
	if err != nil {
		return nil, err
	}

	return t.read()
}

// Get ...
func (t *ClientTxn) Get(key []byte) ([]byte, error) {
	return t.request(actionGet, &key)
}

// Set ...
func (t *ClientTxn) Set(key, value []byte) error {
	_, err := t.request(actionSet, &key, &value)
	return err
}

// Delete ...
func (t *ClientTxn) Delete(key []byte) error {
	_, err := t.request(actionDelete, &key)
	return err
}

// Do ...
func (t *ClientTxn) Do(reverse bool, from []byte, fn kvstore.Iteratee) error {
	key, err := t.request(actionIterate, boolToBytes(reverse), &from)
	if err != nil {
		return err
	}

	for len(key) > 0 {
		err = fn(key)
		if err != nil {
			break
		}

		key, err = t.request(actionIterateNext)
		if err != nil {
			break
		}
	}

	if err == kvstore.ErrStop {
		return nil
	}
	return err
}
