package main

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	Id       int64
	Persona  string
	Username string
	RealName string
	Tm       time.Time
}

func CreatePersonaUser(persona string) (*User, error) {
	user := &User{
		Persona: persona,
		Tm:      time.Now(),
	}

	err := Db.QueryRow(
		`
		INSERT INTO users(persona, tm) 
		VALUES ($1, $2)
		RETURNING id
		`,
		user.Persona, user.Tm,
	).Scan(&user.Id)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func getUserByKey(key string, val interface{}) (*User, error) {
	user := &User{}

	query := fmt.Sprintf(
		`
		SELECT id, COALESCE(persona, ''), COALESCE(username, ''), realname, tm
		FROM users
		WHERE %s = $1
		`,
		key,
	)

	err := Db.QueryRow(query, val).Scan(
		&user.Id, &user.Persona, &user.Username, &user.RealName, &user.Tm,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("User not found")
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserById(id int64) (*User, error) {
	user, err := getUserByKey("id", id)
	return user, err
}

func GetUserByPersona(persona string) (*User, error) {
	user, err := getUserByKey("persona", persona)
	return user, err
}

func GetUserByUsername(username string) (*User, error) {
	user, err := getUserByKey("username", username)
	return user, err
}

func UpdateUser(user *User) error {
	result, err := Db.Exec(
		`
		UPDATE users 
		SET persona = NULLIF($1, ''), username = NULLIF($2, ''), realname = $3, tm = $4
		WHERE id = $5
		`,
		user.Persona, user.Username, user.RealName, user.Tm, user.Id,
	)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return fmt.Errorf("User not found: %d", user.Id)
	}

	return nil
}
