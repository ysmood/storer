package storer_test

import (
	"encoding/binary"

	"github.com/ysmood/byframe"
	"github.com/ysmood/kit"
	"github.com/ysmood/storer"
)

type User struct {
	Name  string
	Level int
}

var _ storer.Unique = &User{}
var _ storer.UniqueType = &User{}
var _ storer.Encoding = &User{}

var UserIDCounter = uint64(0)

func (u *User) ID() []byte {
	b := make([]byte, 8)
	n := binary.PutUvarint(b, UserIDCounter)
	UserIDCounter++
	return b[:n]
}

func (u *User) TypeID() string {
	return kit.RandString(10)
}

func (u *User) Encode() ([]byte, error) {
	n := []byte(u.Name)

	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l, uint64(u.Level))
	return byframe.EncodeTuple(&n, &l), nil
}

func (u *User) Decode(data []byte) error {
	var name, level []byte
	kit.E(byframe.DecodeTuple(data, &name, &level))

	u.Level = int(binary.LittleEndian.Uint64(level))
	u.Name = string(name)
	return nil
}
