package storer_test

import (
	"fmt"

	"github.com/ysmood/storer"
)

func ExampleList_create_read_update_delete() {
	type Person struct {
		Name  string
		Level int
	}

	store := storer.New("")

	users := store.List(&Person{})

	// add
	id, _ := users.Add(&Person{"Jack", 20})

	// get
	var jack Person
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
	type Account struct {
		Name  string
		Level int
	}

	store := storer.New("")

	users := store.List(&Account{})

	// setup index
	index := users.Index("level", func(u *Account) interface{} {
		return u.Level
	})

	// add test data
	_, _ = users.Add(&Account{"A", 1})
	_, _ = users.Add(&Account{"B", 2})
	_, _ = users.Add(&Account{"C", 2})
	_, _ = users.Add(&Account{"D", 3})
	_, _ = users.Add(&Account{"E", 5})

	// get the first user whose level is 1
	var user3 Account
	_ = index.From(3).Find(&user3)
	fmt.Println(user3)

	// get users whose level is 2
	var users2 []Account
	_ = index.From(2).Find(&users2)
	fmt.Println(len(users2))

	// get users whose level is between 3 and 6
	var twenties []Account
	_ = index.From(3).Filter(&twenties, func(u *Account) bool {
		return u.Level < 6
	})
	fmt.Println(twenties)

	// get users whose level is odd
	var odd []Account
	_ = index.From(nil).Filter(&odd, func(u *Account) bool {
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
	type Level struct {
		Num int
	}

	store := storer.New("")

	users := store.List(&Level{})

	// add an item
	id, _ := users.Add(&Level{1})

	// atomic level up
	_ = store.Update(func(txn storer.Txn) error {
		usersTxn := users.Txn(txn)

		// get item
		var user Level
		_ = usersTxn.Get(id, &user)

		// level up
		user.Num++

		// update item
		_ = usersTxn.Set(id, &user)

		return nil
	})

	// check the update
	var user Level
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
	config := store.Value("", &Config{"test.com", 8080})

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
