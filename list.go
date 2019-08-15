package storer

import (
	"errors"
	"reflect"

	"github.com/ysmood/storer/pkg/kvstore"
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
	id := GenID(item)
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
	return listTxn.dictTxn.GetByBytes(id, item)
}

// SetByBytes update an existing item
func (listTxn *ListTxn) SetByBytes(id []byte, item interface{}) error {
	oldItem := reflect.New(listTxn.list.dict.itemType.Elem())

	// if the item doesn't exist return error
	err := listTxn.GetByBytes(id, oldItem.Interface())
	if err != nil {
		return err
	}

	err = listTxn.dictTxn.SetByBytes(id, item)
	if err != nil {
		return err
	}

	for _, index := range listTxn.list.indexes {
		err = index.update(listTxn.dictTxn.txn, id, oldItem.Interface(), item)
		if err != nil {
			return err
		}
	}

	return nil
}

// DelByBytes remove a item from the list
func (listTxn *ListTxn) DelByBytes(id []byte) error {
	item := reflect.New(listTxn.list.dict.itemType.Elem())
	err := listTxn.GetByBytes(id, item.Interface())
	if err != nil {
		return err
	}

	for _, index := range listTxn.list.indexes {
		err := index.del(listTxn.dictTxn.txn, id, item.Interface())
		if err != nil {
			return err
		}
	}

	return listTxn.dictTxn.DelByBytes(id)
}

// IndexByBytes byte version of Index
func (listTxn *ListTxn) IndexByBytes(name string, fn GenIndexBytes) (*Index, error) {
	bucket, err := listTxn.list.dict.store.bucket(indexType, listTxn.list.dict.name, name)
	if err != nil {
		return nil, err
	}

	_, has := listTxn.list.indexes[name]
	if has {
		return nil, ErrIndexExists
	}

	index := &Index{
		name:     name,
		list:     listTxn.list,
		bucket:   bucket,
		genIndex: fn,
	}

	listTxn.list.indexes[name] = index

	return index, nil
}

// Each ...
func (listTxn *ListTxn) Each(fn MapEach) error {
	return listTxn.dictTxn.Each(fn)
}
