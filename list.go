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
	m       *Map
	indexes map[string]*Index
}

// ListTxn ...
type ListTxn struct {
	list   *List
	mapTxn *MapTxn
}

// Txn create transaction context
func (list *List) Txn(txn kvstore.Txn) *ListTxn {
	return &ListTxn{
		list:   list,
		mapTxn: list.m.Txn(txn),
	}
}

// AddByBytes add an item to the list, return the id and error
func (t *ListTxn) AddByBytes(item interface{}) ([]byte, error) {
	id := id(item)
	err := t.mapTxn.SetByBytes(id, item)
	if err != nil {
		return nil, err
	}

	for _, index := range t.list.indexes {
		err = index.add(t.mapTxn.txn, id, item)
		if err != nil {
			return nil, err
		}
	}

	return id, nil
}

// GetByBytes get item from the list
func (t *ListTxn) GetByBytes(id []byte, item interface{}) error {
	return t.mapTxn.GetByBytes(id, item)
}

// SetByBytes update an existing item
func (t *ListTxn) SetByBytes(id []byte, item interface{}) error {
	oldItem := reflect.New(t.list.m.itemType.Elem())

	// if the item doesn't exist return error
	err := t.GetByBytes(id, oldItem.Interface())
	if err != nil {
		return err
	}

	err = t.mapTxn.SetByBytes(id, item)
	if err != nil {
		return err
	}

	for _, index := range t.list.indexes {
		err = index.update(t.mapTxn.txn, id, oldItem.Interface(), item)
		if err != nil {
			return err
		}
	}

	return nil
}

// DelByBytes remove a item from the list
func (t *ListTxn) DelByBytes(id []byte) error {
	item := reflect.New(t.list.m.itemType.Elem())
	err := t.GetByBytes(id, item.Interface())
	if err != nil {
		return err
	}

	for _, index := range t.list.indexes {
		err := index.del(t.mapTxn.txn, id, item.Interface())
		if err != nil {
			return err
		}
	}

	return t.mapTxn.DelByBytes(id)
}

// IndexByBytes byte version of Index
func (t *ListTxn) IndexByBytes(name string, fn GenIndexBytes) (*Index, error) {
	bucket, err := t.list.m.store.bucket(indexType, t.list.m.name, name)
	if err != nil {
		return nil, err
	}

	_, has := t.list.indexes[name]
	if has {
		return nil, ErrIndexExists
	}

	index := &Index{
		name:     name,
		list:     t.list,
		bucket:   bucket,
		genIndex: fn,
	}

	t.list.indexes[name] = index

	return index, nil
}
