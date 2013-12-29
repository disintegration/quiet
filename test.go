package main

import (
	"fmt"
	"log"
)

func Test() {
	var err error

	for i := 0; i < 10; i++ {
		CreatePersonaUser(
			fmt.Sprintf("test%d@example.com", i),
		)
	}

	users := make([]*User, 10)

	for i := 0; i < 10; i++ {
		users[i], err = GetUserById(int64(i + 1))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User%d = %+v\n", i, users[i])
	}

	for i := 0; i < 10; i++ {
		users[i].Username = fmt.Sprintf("username%d", i)
		err = UpdateUser(users[i])
		if err != nil {
			log.Fatal(err)
		}

		u, err := GetUserByUsername(fmt.Sprintf("username%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User%d = %+v\n", i, u)
	}

	EnqueueAvatar(users[0])
	EnqueueAvatar(users[1])
	EnqueueAvatar(users[2])

}
