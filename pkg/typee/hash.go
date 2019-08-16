package typee

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// String as long as the passed in value is the same type
// same string will be returned.
// The order of the struct fields doesn't matter.
func String(t reflect.Type) string {
	return toString(t, 0)
}

func newline(level int) string {
	s := "\n"
	for i := 0; i < level; i++ {
		s += "    "
	}
	return s
}

func toString(t reflect.Type, level int) string {
	switch t.Kind() {
	case reflect.Struct:
		l := t.NumField()
		types := []string{}
		for i := 0; i < l; i++ {
			field := t.Field(i)
			str := toString(field.Type, level+1)
			types = append(types, field.Name+" "+str)
		}
		sort.Strings(types)
		return "{" + newline(level+1) + strings.Join(types, newline(level+1)) + newline(level) + "}"
	case reflect.Slice:
		return toString(t.Elem(), level)
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), toString(t.Elem(), level))
	case reflect.Map:
		str := toString(t.Elem(), level)
		return fmt.Sprintf("[%s]%s", toString(t.Key(), level), str)
	default:
		return t.String()
	}
}

// Hash get hash value of the type
func Hash(types ...reflect.Type) []byte {
	s := ""
	for _, t := range types {
		s += String(t)
	}

	// md5 should be enough just for unintentional inputs
	h := md5.Sum([]byte(s))
	return h[:]
}
