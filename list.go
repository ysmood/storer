package storer

import (
	"errors"
	"reflect"

	"github.com/ysmood/storer/pkg/kvstore"
	"github.com/ysmood/storer/pkg/typee"
)

// ErrIndexExists ...
var ErrIndexExists = errors.New("index already exists")

// List ...
type List struct {
	dict    *Map
	indexes map[string]*Index
}

// ListTxn ...
type ListTxn struct {
	list    *List
	dictTxn *MapTxn
}

// Txn create transaction context
func (list *List) Txn(txn kvstore.Txn) *ListTxn {
	return &ListTxn{
		list:    list,
		dictTxn: list.dict.Txn(txn),
	}
}

// AddByBytes add an item to the list, return the id and error
func (listTxn *ListTxn) AddByBytes(item interface{}) ([]byte, error) {
	id := typee.GenID(item)
	err := listTxn.dictTxn.SetByBytes(id, item)
	if err != nil {
		return nil, err
	}

	for _, index := range listTxn.list.indexes {
		err = index.add(listTxn.dictTxn.txn, id, item)
		if err != nil {
			return nil, err
		}
	}

	return id, nil
}

// GetByBytes get item from the list
func (listTxn *ListTxn) GetByBytes(id []byte, item interface{}) error {
	err := listTxn.dictTxn.GetByBytes(id, item)
	if err == typee.ErrMigrated {
		return listTxn.updateIndex(id, item)
	}
	return err
}

// SetByBytes update an existing item
func (listTxn *ListTxn) SetByBytes(id []byte, item interface{}) error {
	err := listTxn.dictTxn.SetByBytes(id, item)
	if err != nil {
		return err
	}

	return listTxn.updateIndex(id, item)
}

func (listTxn *ListTxn) updateIndex(id []byte, item interface{}) error {
	for _, index := range listTxn.list.indexes {
		err := index.update(listTxn.dictTxn.txn, id, item)
		if err != nil {
			return err
		}
	}
	return nil
}

// DelByBytes remove a item from the list
func (listTxn *ListTxn) DelByBytes(id []byte) error {
	item := reflect.New(listTxn.list.dict.typeID.Type)
	err := listTxn.GetByBytes(id, item.Interface())
	if err != nil {
		return err
	}

	for _, index := range listTxn.list.indexes {
		err := index.del(listTxn.dictTxn.txn, id)
		if err != nil {
			return err
		}
	}

	return listTxn.dictTxn.DelByBytes(id)
}

// IndexByBytes byte version of Index
func (listTxn *ListTxn) IndexByBytes(name string, fn GenIndexBytes) (*Index, error) {
	_, has := listTxn.list.indexes[name]
	if has {
		return nil, ErrIndexExists
	}

	index := &Index{
		name: name,
		list: listTxn.list,
		bucket: listTxn.list.dict.store.bucket(
			listTxn.list.dict.typeID.Anchor,
			listTxn.list.dict.name,
			"index",
			name,
		),
		rbucket: listTxn.list.dict.store.bucket(
			listTxn.list.dict.typeID.Anchor,
			listTxn.list.dict.name,
			"rindex",
			name,
		),
		genIndex: fn,
	}

	listTxn.list.indexes[name] = index

	return index, nil
}

// Each ...
func (listTxn *ListTxn) Each(fn MapEach) error {
	return listTxn.dictTxn.Each(fn)
}
