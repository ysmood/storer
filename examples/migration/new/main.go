package main

import (
	"fmt"

	"github.com/ysmood/storer"
	"github.com/ysmood/storer/pkg/typee"
)

// PersonV0 the old type to be migrated
type PersonV0 struct {
	// FullName ...
	FullName string
}

// Person the new type to migrate to has an extra field NameLen
type Person struct {
	// Name ...
	Name string
	// NameLen the length of the name
	NameLen int
}

var _ typee.Migratable = &Person{}

// Precedent ...
func (p *Person) Precedent() interface{} {
	return &PersonV0{}
}

// Migrate ...
func (p *Person) Migrate(item interface{}) {
	old := item.(*PersonV0)
	p.Name = old.FullName
	p.NameLen = len(old.FullName)
}

func main() {
	store := storer.New("tmp")

	m := store.Map(&Person{})

	var jack Person
	_ = m.Get("1", &jack)

	fmt.Println(jack.Name, jack.NameLen)
}
