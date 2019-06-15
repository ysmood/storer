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
