package typee_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/storer/pkg/typee"
)

type TestType struct {
	String string
	Int    int
}

func TestGenTypeID(t *testing.T) {
	hash := typee.GenTypeID(&TestType{}).ID
	assert.Equal(t, "04182418b3f271d3b24b75da9902a174", fmt.Sprintf("%x", hash))
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
