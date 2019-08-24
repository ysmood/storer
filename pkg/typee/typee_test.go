package typee_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/storer/pkg/typee"
)

func TestGenTypeID(t *testing.T) {
	type TestType struct {
		String string
		Int    int
	}

	hash := typee.GenTypeID(&TestType{}).ID
	assert.Equal(t, "04182418b3f271d3b24b75da9902a174", fmt.Sprintf("%x", hash))

	id := typee.GenID(&TestType{})
	assert.Len(t, id, 12)
}

type TestGenIDType struct {
}

var _ typee.Unique = &TestGenIDType{}

func (t *TestGenIDType) UUID() []byte {
	return []byte("ok")
}

func TestGenID(t *testing.T) {
	id := typee.GenID(&TestGenIDType{})
	assert.Equal(t, []byte("ok"), id)
}

type UserV0 struct {
	Int int
}

type UserV1 struct {
	Str string
}

var _ typee.Migratable = &UserV1{}

func (u *UserV1) Precedent() interface{} { return &UserV0{} }

func (u *UserV1) Migrate(item interface{}) {
	old := item.(*UserV0)
	u.Str = fmt.Sprint(old.Int)
}

type User struct {
	Name string
}

var _ typee.Migratable = &User{}

func (u *User) Precedent() interface{} { return &UserV1{} }

func (u *User) Migrate(item interface{}) {
	old := item.(*UserV1)
	u.Name = old.Str
}

func TestMigration(t *testing.T) {
	data0, _ := typee.Encode(&UserV0{0}, nil)
	data1, _ := typee.Encode(&UserV1{"1"}, nil)

	var v1From0 UserV1
	err := typee.Decode(data0, &v1From0, nil)
	assert.Equal(t, "0", v1From0.Str)
	assert.Equal(t, typee.ErrMigrated, err)

	var vFrom0 User
	_ = typee.Decode(data0, &vFrom0, nil)
	assert.Equal(t, "0", vFrom0.Name)

	var vFrom1 User
	_ = typee.Decode(data1, &vFrom1, nil)
	assert.Equal(t, "1", vFrom1.Name)
}

type EncodeErr struct {
	err error
}

func (e *EncodeErr) Encode() ([]byte, error) {
	return nil, e.err
}

func (e *EncodeErr) Decode([]byte) error {
	return e.err
}

var _ typee.Encoding = &EncodeErr{}

func TestEncodeErr(t *testing.T) {
	err := errors.New("err")
	_, e := typee.Encode(&EncodeErr{}, nil)
	assert.Equal(t, err, e)

	var item EncodeErr
	e = typee.Decode([]byte{}, &item, nil)
	assert.Equal(t, err, e)
}
