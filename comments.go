package main

import (
	"database/sql"
	"fmt"
	"time"
)

type Comment struct {
	Id      int64
	UserId  int64
	PhotoId int64
	Comment string
	Tm      time.Time

	UserUsername string
	UserRealName string
}

func CreateComment(userId int64, photoId int64, comment string) (*Comment, error) {
	cmt := &Comment{
		UserId:  userId,
		PhotoId: photoId,
		Comment: comment,
		Tm:      time.Now(),
	}

	err := Db.QueryRow(
		`
		INSERT INTO comments(user_id, photo_id, comment, tm) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`,
		cmt.UserId, cmt.PhotoId, cmt.Comment, cmt.Tm,
	).Scan(&cmt.Id)

	if err != nil {
		return nil, err
	}
	return cmt, nil
}

func GetCommentById(id int64) (*Comment, error) {
	cmt := &Comment{Id: id}

	err := Db.QueryRow(
		`
		SELECT c.id, u.id, u.username, u.realname, c.photo_id, c.comment, c.tm
		FROM 
			comments c
			JOIN users u on c.user_id = u.id
		WHERE c.id = $1
		`,
		id,
	).Scan(
		&cmt.Id,
		&cmt.UserId, &cmt.UserUsername, &cmt.UserRealName,
		&cmt.PhotoId, &cmt.Comment, &cmt.Tm,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Comment not found: %d", cmt.Id)
	} else if err != nil {
		return nil, err
	}

	return cmt, nil
}

func GetCommentsByPhotoId(photoId int64) ([]*Comment, error) {
	result := make([]*Comment, 0, 10)

	rows, err := Db.Query(
		`
		SELECT c.id, u.id, u.username, u.realname, c.photo_id, c.comment, c.tm
		FROM 
			comments c
			JOIN users u on c.user_id = u.id
		WHERE c.photo_id = $1
		ORDER BY c.tm
		`,
		photoId,
	)

	if err != nil {
		return []*Comment{}, err
	}

	for rows.Next() {
		cmt := &Comment{}
		err := rows.Scan(
			&cmt.Id,
			&cmt.UserId, &cmt.UserUsername, &cmt.UserRealName,
			&cmt.PhotoId, &cmt.Comment, &cmt.Tm,
		)
		if err != nil {
			return []*Comment{}, err
		}
		result = append(result, cmt)
	}

	if err := rows.Err(); err != nil {
		return []*Comment{}, err
	}

	return result, nil
}

func DelCommentById(id int64) error {
	_, err := Db.Exec(`DELETE FROM comments WHERE id = $1`, id)
	return err
}
