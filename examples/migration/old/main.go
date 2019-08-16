package main

import (
	"github.com/ysmood/storer"
)

// Person the old type to be migrated
type Person struct {
	// FullName ...
	FullName string
}

func main() {
	store := storer.New("tmp")

	m := store.Map(&Person{})
	_ = m.Set("1", &Person{"jack"})
}
