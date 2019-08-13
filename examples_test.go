package storer_test

import (
	"fmt"

	"github.com/ysmood/storer"
)

func ExampleStore_create_read_update_delete() {
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

func ExampleStore_indexing() {
	type User struct {
		Name  string
		Level int
	}

	store := storer.New("")

	users := store.List(&User{})

	// setup index
	index := users.Index("level", func(ctx *storer.GenCtx) interface{} {
		return ctx.Item.(*User).Level
	})

	// add test data
	_, _ = users.Add(&User{"A", 1})
	_, _ = users.Add(&User{"B", 2})
	_, _ = users.Add(&User{"C", 2})
	_, _ = users.Add(&User{"D", 3})
	_, _ = users.Add(&User{"E", 5})

	// get the first user who is 1
	var user3 User
	_ = index.From(3).Find(&user3)
	fmt.Println(user3)

	// get users who is 2
	var users2 []User
	_ = index.From(2).Find(&users2)
	fmt.Println(len(users2))

	// get users between 3 and 6
	var twenties []User
	_ = index.From(3).Filter(&twenties, func(ctx *storer.IterCtx) (bool, bool) {
		match := ctx.Compare(6) < 0
		isContinue := match
		return match, isContinue
	})
	fmt.Println(twenties)

	// get users whose level are odd
	var even []User
	_ = index.From(nil).Filter(&even, func(ctx *storer.IterCtx) (bool, bool) {
		var user User
		_ = ctx.Item(&user)

		return user.Level%2 == 1, true
	})
	fmt.Println(even)

	// Output:
	// {D 3}
	// 2
	// [{D 3} {E 5}]
	// [{A 1} {D 3} {E 5}]
}

func ExampleStore_transaction() {
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
