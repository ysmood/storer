package typee

import (
	"bytes"
	"crypto/rand"
	"errors"
	"reflect"
	"sync"

	"github.com/vmihailenco/msgpack"
	"github.com/ysmood/byframe"
)

// Unique custom id generator, by default crypto safe uid will be used
type Unique interface {
	// UUID should return an unique id for each different instance
	UUID() []byte
}

// TypeAnchorable ...
type TypeAnchorable interface {
	// TypeAnchor should return the same value after type change
	TypeAnchor() string
}

// Encoding custom encoding handler, by default msgpack is used
type Encoding interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

// Migratable ...
type Migratable interface {
	// Precedent return previous type element
	Precedent() interface{}
	// Migrate method used to upgrade type
	Migrate(item interface{})
}

// GenID generate an unique id
func GenID(val interface{}) []byte {
	id, ok := val.(Unique)
	if ok {
		return id.UUID()
	}
	return randBytes(12)
}

// Encode ...
func Encode(item interface{}, mapper Mapper) (data []byte, err error) {
	encoding, ok := item.(Encoding)

	if ok {
		data, err = encoding.Encode()
	} else {
		data, err = msgpack.Marshal(item)
	}

	if err != nil {
		return nil, err
	}

	if mapper == nil {
		mapper = defaultMapper
	}

	version, err := mapper(GenTypeID(item).ID)
	if err != nil {
		return nil, err
	}

	return byframe.EncodeTuple(&version, &data), nil
}

// ErrMigrated ...
var ErrMigrated = errors.New("[storer.typee] migrated")

// Decode when data is migrated ErrMigrated will be returned
func Decode(versioned []byte, item interface{}, mapper Mapper) error {
	var version, data []byte

	err := byframe.DecodeTuple(versioned, &version, &data)
	if err != nil {
		return err
	}

	tasks, item, err := migrateTasks(item, version, mapper)
	if err != nil {
		return err
	}

	encoding, ok := item.(Encoding)
	if ok {
		err = encoding.Decode(data)
	} else {
		err = msgpack.Unmarshal(data, item)
	}
	if err != nil {
		return err
	}

	for i := len(tasks) - 1; i >= 0; i-- {
		elem := tasks[i]

		elem.Migrate(item)

		item = elem
	}

	if len(tasks) > 0 {
		return ErrMigrated
	}

	return nil
}

// ErrNotMigratable ...
var ErrNotMigratable = errors.New("[storer.typee] item must implement Migratable interface")

func migrateTasks(item interface{}, version []byte, mapper Mapper) ([]Migratable, interface{}, error) {
	if mapper == nil {
		mapper = defaultMapper
	}

	list := []Migratable{}

	for {
		itemVersion, err := mapper(GenTypeID(item).ID)
		if err != nil {
			return nil, nil, err
		}

		if bytes.Equal(version, itemVersion) {
			return list, item, nil
		}

		m, ok := item.(Migratable)
		if !ok {
			return nil, nil, ErrNotMigratable
		}
		list = append(list, m)

		item = m.Precedent()
	}
}

// Mapper persistently map long bytes to short bytes.
// Same long id should always map to the same short id after restart the program.
// If the mapper is nil, long id will be used
type Mapper func(longID []byte) (shortID []byte, err error)

func defaultMapper(longID []byte) ([]byte, error) {
	return longID, nil
}

// because a type cannot change on the fly, so it's safe to cache the hash of types
var typeIDCache = &sync.Map{}

// TypeID ...
type TypeID struct {
	// ID ...
	ID []byte
	// Anchor ...
	Anchor string
	// Type ...
	Type reflect.Type
}

// GenTypeID generate unique id from type history
// return typeAnchor and typeID
func GenTypeID(item interface{}) *TypeID {
	t := getElemType(item)

	var anchor string
	if anchorable, ok := item.(TypeAnchorable); ok {
		anchor = anchorable.TypeAnchor()
	} else {
		anchor = t.String()
	}

	if typeID, ok := typeIDCache.Load(anchor); ok {
		return typeID.(*TypeID)
	}

	history := []reflect.Type{t}

	m, ok := item.(Migratable)
	for ok {
		pre := m.Precedent()
		history = append(history, getElemType(pre))
		m, ok = pre.(Migratable)
	}

	typeID := &TypeID{
		ID:     Hash(history...),
		Anchor: anchor,
		Type:   t,
	}
	typeIDCache.Store(anchor, typeID)
	return typeID
}

// ErrNotPtr ...
var ErrNotPtr = errors.New("[storer.typee] must be a pointer to the item")

func getElemType(item interface{}) reflect.Type {
	t := reflect.TypeOf(item)
	if t.Kind() != reflect.Ptr {
		panic(ErrNotPtr)
	}
	return t.Elem()
}

func randBytes(len int) []byte {
	b := make([]byte, len)
	_, _ = rand.Read(b)
	return b
}
