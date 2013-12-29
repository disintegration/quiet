package main

import (
	"time"
)

type Contact struct {
	Id        int64
	UserId    int64
	ContactId int64
	Tm        time.Time

	UserUsername    string
	UserRealName    string
	ContactUsername string
	ContactRealName string
}

func CreateContact(userId int64, contactId int64) (*Contact, error) {
	cnt := &Contact{
		UserId:    userId,
		ContactId: contactId,
		Tm:        time.Now(),
	}

	err := Db.QueryRow(
		`
		INSERT INTO contacts(user_id, contact_id, tm) 
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		cnt.UserId, cnt.ContactId, cnt.Tm,
	).Scan(&cnt.Id)

	if err != nil {
		return nil, err
	}
	return cnt, nil
}

func IsContacted(userId int64, contactId int64) (bool, error) {
	var count int

	err := Db.QueryRow(
		`SELECT COUNT(*) FROM contacts WHERE user_id = $1 AND contact_id = $2`,
		userId, contactId,
	).Scan(&count)

	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func DelContact(userId int64, contactId int64) error {
	_, err := Db.Exec(`DELETE FROM contacts WHERE user_id = $1 AND contact_id = $2`, userId, contactId)
	return err
}

func GetContactsByUserId(id int64) ([]*Contact, error) {
	result := make([]*Contact, 0, 1)

	rows, err := Db.Query(
		`
		SELECT 
			c.id, c.tm,
			u1.id, COALESCE(u1.username, ''), u1.realname,
			u2.id, COALESCE(u2.username, ''), u2.realname 
		FROM 
			users u1,
			contacts c,
			users u2
		WHERE 
			u1.id=$1
			AND c.user_id = u1.id
			AND c.contact_id = u2.id
		ORDER BY u2.username
		`,
		id,
	)

	if err != nil {
		return []*Contact{}, err
	}

	for rows.Next() {
		contact := &Contact{}
		err := rows.Scan(
			&contact.Id, &contact.Tm,
			&contact.UserId, &contact.UserUsername, &contact.UserRealName,
			&contact.ContactId, &contact.ContactUsername, &contact.ContactRealName,
		)
		if err != nil {
			return []*Contact{}, err
		}
		result = append(result, contact)
	}

	if err := rows.Err(); err != nil {
		return []*Contact{}, err
	}

	return result, nil
}
