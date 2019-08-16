package typee_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysmood/storer/pkg/typee"
)

const fixture01 = `{
    A {
        Array [5]{
            Int int
        }
        Int int
        Map [{
            String string
        }]{
            Int int
        }
        Slice {
            Int int
        }
    }
    String string
}`

const fixture02 = "60fabadae6a7ee0339728d93b2d40341"

func TestHash(t *testing.T) {
	type Key struct {
		String string
	}
	type Value struct {
		Int int
	}

	type A struct {
		Int   int
		Slice []Value
		Array [5]Value
		Map   map[Key]Value
	}

	type B struct {
		String string
		A      A
	}

	p := reflect.TypeOf(B{})

	s := typee.String(p)

	assert.Equal(t, fixture01, s)

	assert.Equal(t, fixture02, fmt.Sprintf("%x", typee.Hash(p)))
}
