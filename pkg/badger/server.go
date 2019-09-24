package badger

import (
	"errors"
	"net"

	"github.com/ysmood/byframe"
	"github.com/ysmood/storer/pkg/kvstore"
)

type actionEnum int

const (
	actionBegin actionEnum = iota
	actionGet
	actionSet
	actionDelete
	actionIterate
	actionIterateNext
	actionEnd
)

var (
	emptyBytes = []byte{}
	falseBytes = []byte{0}
	trueBytes  = []byte{1}

	actionBeginBytes       = []byte{byte(actionBegin)}
	actionGetBytes         = []byte{byte(actionGet)}
	actionSetBytes         = []byte{byte(actionSet)}
	actionDeleteBytes      = []byte{byte(actionDelete)}
	actionIterateBytes     = []byte{byte(actionIterate)}
	actionIterateNextBytes = []byte{byte(actionIterateNext)}
	actionEndBytes         = []byte{byte(actionEnd)}
)

func (a actionEnum) bytes() *[]byte {
	switch a {
	case actionBegin:
		return &actionBeginBytes
	case actionGet:
		return &actionGetBytes
	case actionSet:
		return &actionSetBytes
	case actionDelete:
		return &actionDeleteBytes
	case actionIterate:
		return &actionIterateBytes
	case actionIterateNext:
		return &actionIterateNextBytes
	case actionEnd:
		return &actionEndBytes
	default:
		panic("undefined action")
	}
}

func bytesToAction(b []byte) actionEnum {
	return actionEnum(b[0])
}

type serverTxn struct {
	conn    net.Conn
	db      *Badger
	scanner *byframe.Scanner
}

// Serve ...
func Serve(db *Badger, l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		txn := serverTxn{
			conn:    c,
			db:      db,
			scanner: byframe.NewScanner(c),
		}

		go txn.handleConn()
	}
}

func (txn *serverTxn) handleConn() {
	action, update, err := txn.read()
	if action != actionBegin {
		txn.response(nil, errors.New("must begin with actionBegin"))
		return
	}
	if err != nil {
		txn.response(nil, err)
		return
	}
	txn.response(nil, nil)

	txnErr := txn.db.Do(bytesToBool(update), txn.do)

	action, _, err = txn.read()
	if action != actionEnd {
		txn.response(nil, errors.New("must end with actionEnd"))
		return
	}
	if err != nil {
		txn.response(nil, err)
		return
	}
	txn.response(nil, txnErr)

	_ = txn.conn.Close()
}

func (txn *serverTxn) do(t kvstore.Txn) error {
	for {
		action, args, err := txn.read()
		if err != nil {
			return err
		}

		switch action {
		case actionGet:
			key := args
			val, err := t.Get(key)
			if err != nil {
				txn.response(nil, err)
				continue
			}
			txn.response(val, nil)

		case actionSet:
			var key, val []byte
			err := byframe.DecodeTuple(args, &key, &val)
			if err != nil {
				txn.response(nil, err)
				continue
			}
			txn.response(nil, t.Set(key, val))

		case actionDelete:
			key := args
			txn.response(nil, t.Delete(key))

		case actionIterate:
			var reverse, from []byte
			err := byframe.DecodeTuple(args, &reverse, &from)
			if err != nil {
				txn.response(nil, err)
				continue
			}

			txn.response(nil, t.Do(bytesToBool(reverse), from, func(key []byte) error {
				txn.response(key, nil)
				return txn.do(t)
			}))

		case actionIterateNext:
			return nil

		case actionEnd:
			txn.response(nil, nil)
			return nil

		default:
			return errors.New("unknown action")
		}
	}
}

func (txn *serverTxn) read() (actionEnum, []byte, error) {
	if !txn.scanner.Scan() {
		err := txn.scanner.Err()
		if err != nil {
			return actionEnd, nil, err
		}
	}

	var action, args []byte
	err := byframe.DecodeTuple(txn.scanner.Frame(), &action, &args)

	return bytesToAction(action), args, err
}

func (txn *serverTxn) response(data []byte, err error) {
	var errData []byte
	if err != nil {
		errData = []byte(err.Error())
	}

	_, e := txn.conn.Write(byframe.Encode(byframe.EncodeTuple(&data, &errData)))
	if e != nil {
		panic(e)
	}
}
