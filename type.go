package storer

import (
	"crypto/rand"
	"errors"
	"reflect"

	"github.com/vmihailenco/msgpack"
)

// ErrItemType ...
var ErrItemType = errors.New("wrong item type")

const listType = "list"
const mapType = "map"
const indexType = "index"

// Unique custom id generator, by default it will use a crypto safe uid
type Unique interface {
	// ID should return an unique id for its scope
	ID() []byte
}

// UniqueType custom type generator, by default reflection name will be used
type UniqueType interface {
	TypeID() string
}

// Encoding custom encoding handler, by default msgpack is used
type Encoding interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

func id(val interface{}) []byte {
	id, ok := val.(Unique)
	if ok {
		return id.ID()
	}
	return randBytes(12)
}

// Encode ...
func Encode(item interface{}) ([]byte, error) {
	encoding, ok := item.(Encoding)
	if ok {
		return encoding.Encode(item)
	}

	return msgpack.Marshal(item)
}

// Decode ...
func Decode(data []byte, item interface{}) error {
	encoding, ok := item.(Encoding)
	if ok {
		return encoding.Decode(data, item)
	}

	return msgpack.Unmarshal(data, item)
}

// generate unique id from type reflection
func genTypeID(item interface{}) (reflect.Type, string) {
	t := reflect.TypeOf(item)

	if uniqueType, ok := item.(UniqueType); ok {
		return t, uniqueType.TypeID()
	}

	return t, t.String()
}

func randBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = rand.Read(b)
	return b
}
