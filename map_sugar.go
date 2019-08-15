package storer

// Set string version of MapTxn.SetByBytes
func (t *MapTxn) Set(id string, item interface{}) error {
	return t.SetByBytes([]byte(id), item)
}

// Get string version of MapTxn.GetByBytes
func (t *MapTxn) Get(id string, item interface{}) error {
	return t.GetByBytes([]byte(id), item)
}

// Del string version of MapTxn.DelByBytes
func (t *MapTxn) Del(id string) error {
	return t.DelByBytes([]byte(id))
}

// SetByBytes ...
func (m *Map) SetByBytes(id []byte, item interface{}) error {
	return m.store.Update(func(txn Txn) error {
		return m.Txn(txn).SetByBytes(id, item)
	})
}

// GetByBytes ...
func (m *Map) GetByBytes(id []byte, item interface{}) error {
	return m.store.View(func(txn Txn) error {
		return m.Txn(txn).GetByBytes(id, item)
	})
}

// DelByBytes ...
func (m *Map) DelByBytes(id []byte) error {
	return m.store.Update(func(txn Txn) error {
		return m.Txn(txn).DelByBytes([]byte(id))
	})
}

// Set ...
func (m *Map) Set(id string, item interface{}) error {
	return m.SetByBytes([]byte(id), item)
}

// Get ...
func (m *Map) Get(id string, item interface{}) error {
	return m.GetByBytes([]byte(id), item)
}

// Del ...
func (m *Map) Del(id string) error {
	return m.DelByBytes([]byte(id))
}
