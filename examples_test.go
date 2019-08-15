package storer_test

import (
	"fmt"

	"github.com/ysmood/storer"
)

func ExampleList_create_read_update_delete() {
	type User struct {
		Name  string
		Level int
	}

	store := storer.New("")

	users := store.List(&User{})

	// add
	id, _ := users.Add(&User{"Jack", 20})

	// get
	var jack User
	_ = users.Get(id, &jack)
	fmt.Println(jack)

	// update
	jack.Level = 21
	_ = users.Set(id, &jack)

	// delete
	_ = users.Del(id)

	// Output: {Jack 20}
}

func ExampleList_indexing() {
	type User struct {
		Name  string
		Level int
	}

	store := storer.New("")

	users := store.List(&User{})

	// setup index
	index := users.Index("level", func(u *User) interface{} {
		return u.Level
	})

	// add test data
	_, _ = users.Add(&User{"A", 1})
	_, _ = users.Add(&User{"B", 2})
	_, _ = users.Add(&User{"C", 2})
	_, _ = users.Add(&User{"D", 3})
	_, _ = users.Add(&User{"E", 5})

	// get the first user whose level is 1
	var user3 User
	_ = index.From(3).Find(&user3)
	fmt.Println(user3)

	// get users whose level is 2
	var users2 []User
	_ = index.From(2).Find(&users2)
	fmt.Println(len(users2))

	// get users whose level is between 3 and 6
	var twenties []User
	_ = index.From(3).Filter(&twenties, func(u *User) bool {
		return u.Level < 6
	})
	fmt.Println(twenties)

	// get users whose level is odd
	var odd []User
	_ = index.From(nil).Filter(&odd, func(u *User) bool {
		return u.Level%2 == 1
	})
	fmt.Println(odd)

	// Output:
	// {D 3}
	// 2
	// [{D 3} {E 5}]
	// [{A 1} {D 3} {E 5}]
}

func ExampleList_transaction() {
	type User struct {
		Level int
	}

	store := storer.New("")

	users := store.List(&User{})

	// add an item
	id, _ := users.Add(&User{1})

	// atomic level up
	_ = store.Update(func(txn storer.Txn) error {
		usersTxn := users.Txn(txn)

		// get item
		var user User
		_ = usersTxn.Get(id, &user)

		// level up
		user.Level++

		// update item
		_ = usersTxn.Set(id, &user)

		return nil
	})

	// check the update
	var user User
	_ = users.Get(id, &user)
	fmt.Println(user)

	// Output: {2}
}

func ExampleValue() {
	type Config struct {
		Host string
		Port int
	}

	store := storer.New("")

	// init value of the config
	config := store.Value(&Config{"test.com", 8080})

	// update config
	var c Config
	_ = config.Get(&c)
	c.Port = 80
	_ = config.Set(&c)

	// get config after restart
	_ = config.Get(&c)
	fmt.Println(c)

	// Output: {test.com 80}
}
