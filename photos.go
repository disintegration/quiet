package main

import (
	"database/sql"
	"fmt"
	"time"
)

type Photo struct {
	Id          int64
	UserId      int64
	RandId      string
	Tm          time.Time
	Processed   int
	Title       string
	Description string
	ViewsCount  int

	UserUsername   string
	UserRealName   string
	CommentsCount  int
	FavoritesCount int
}

func CreatePhoto(userId int64, title string, description string) (*Photo, error) {
	photo := &Photo{
		UserId:      userId,
		RandId:      GetRandId(20),
		Tm:          time.Now(),
		Processed:   0,
		Title:       title,
		Description: description,
		ViewsCount:  0,
	}

	err := Db.QueryRow(
		`
		INSERT INTO photos(user_id, rand_id, tm, processed, title, description, views_count) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
		`,
		photo.UserId, photo.RandId, photo.Tm, photo.Processed,
		photo.Title, photo.Description, photo.ViewsCount,
	).Scan(&photo.Id)

	if err != nil {
		return nil, err
	}
	return photo, nil
}

func GetPhotoById(id int64) (*Photo, error) {
	photo := &Photo{Id: id}

	err := Db.QueryRow(
		`
		SELECT
			u.id,
			COALESCE(u.username, ''),
			u.realname,
			p.rand_id,
			p.tm,
			p.processed,
			p.title,
			p.description,
			p.views_count,
			(SELECT COUNT(*) FROM comments c WHERE c.photo_id = p.id),
			(SELECT COUNT(*) FROM favorites f WHERE f.photo_id = p.id) 
		FROM 
			photos p 
			JOIN users u ON u.id = p.user_id
		WHERE p.id=$1
		`,
		id,
	).Scan(
		&photo.UserId, &photo.UserUsername, &photo.UserRealName,
		&photo.RandId, &photo.Tm, &photo.Processed, &photo.Title,
		&photo.Description, &photo.ViewsCount,
		&photo.CommentsCount, &photo.FavoritesCount,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Photo not found: %d", photo.Id)
	} else if err != nil {
		return nil, err
	}
	return photo, nil
}

func SetPhotoTitle(id int64, title string) error {
	_, err := Db.Exec(`UPDATE photos SET title = $1 WHERE id = $2`, title, id)
	return err
}

func SetPhotoDescription(id int64, description string) error {
	_, err := Db.Exec(`UPDATE photos SET description = $1 WHERE id = $2`, description, id)
	return err
}

func SetPhotoProcessed(id int64, processed int) error {
	_, err := Db.Exec(`UPDATE photos SET processed = $1 WHERE id = $2`, processed, id)
	return err
}

func IncPhotoViewsCount(id int64) error {
	_, err := Db.Exec(`UPDATE photos SET views_count = views_count + 1 WHERE id = $1`, id)
	return err
}

func DelPhotoById(id int64) error {
	_, err := Db.Exec(`DELETE FROM favorites WHERE photo_id = $1`, id)
	if err != nil {
		return err
	}
	_, err = Db.Exec(`DELETE FROM comments WHERE photo_id = $1`, id)
	if err != nil {
		return err
	}
	_, err = Db.Exec(`DELETE FROM photos WHERE id = $1`, id)
	return err
}

func GetPhotosCountByUserId(userId int64) (int, error) {
	var count int

	err := Db.QueryRow(
		`SELECT COUNT(*) FROM photos p WHERE user_id = $1 AND p.processed = 1`,
		userId,
	).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetContactsPhotosCountByUserId(userId int64) (int, error) {
	var count int

	err := Db.QueryRow(
		`
		SELECT COUNT(*) 
		FROM photos p, contacts c
		WHERE 
			c.user_id = $1 
			AND c.contact_id = p.user_id
			AND p.processed = 1
		`,
		userId,
	).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetPhotosByUserId(userId int64, offset int, limit int) ([]*Photo, error) {
	result := make([]*Photo, 0, 1)

	rows, err := Db.Query(
		`
		SELECT
			p.id,
			u.id,
			COALESCE(u.username, ''),
			u.realname,
			p.rand_id,
			p.tm,
			p.processed,
			p.title,
			p.description,
			p.views_count,
			(SELECT COUNT(*) FROM comments c WHERE c.photo_id = p.id),
			(SELECT COUNT(*) FROM favorites f WHERE f.photo_id = p.id) 
		FROM 
			photos p 
			JOIN users u ON u.id=p.user_id
		WHERE 
			p.user_id = $1
			AND p.processed = 1
		ORDER BY p.tm DESC
		OFFSET $2
		LIMIT $3
		`,
		userId, offset, limit,
	)

	if err != nil {
		return []*Photo{}, err
	}

	for rows.Next() {
		photo := &Photo{}
		err := rows.Scan(
			&photo.Id,
			&photo.UserId, &photo.UserUsername, &photo.UserRealName,
			&photo.RandId, &photo.Tm, &photo.Processed, &photo.Title,
			&photo.Description, &photo.ViewsCount,
			&photo.CommentsCount, &photo.FavoritesCount,
		)
		if err != nil {
			return []*Photo{}, err
		}
		result = append(result, photo)
	}

	if err := rows.Err(); err != nil {
		return []*Photo{}, err
	}

	return result, nil
}

func GetContactsPhotosByUserId(userId int64, offset int, limit int) ([]*Photo, error) {
	result := make([]*Photo, 0, 1)

	rows, err := Db.Query(
		`
		SELECT
			p.id,
			u.id,
			COALESCE(u.username, ''),
			u.realname,
			p.rand_id,
			p.tm,
			p.processed,
			p.title,
			p.description,
			p.views_count,
			(SELECT COUNT(*) FROM comments c WHERE c.photo_id=p.id),
			(SELECT COUNT(*) FROM favorites f WHERE f.photo_id=p.id) 
		FROM 
			photos p 
			JOIN users u ON u.id = p.user_id,
			contacts c
		WHERE 
			c.user_id = $1
			AND c.contact_id = p.user_id
			AND p.processed = 1
		ORDER BY p.tm DESC
		OFFSET $2
		LIMIT $3
		`,
		userId, offset, limit,
	)

	if err != nil {
		return []*Photo{}, err
	}

	for rows.Next() {
		photo := &Photo{}
		err := rows.Scan(
			&photo.Id,
			&photo.UserId, &photo.UserUsername, &photo.UserRealName,
			&photo.RandId, &photo.Tm, &photo.Processed, &photo.Title,
			&photo.Description, &photo.ViewsCount,
			&photo.CommentsCount, &photo.FavoritesCount,
		)
		if err != nil {
			return []*Photo{}, err
		}
		result = append(result, photo)
	}

	if err := rows.Err(); err != nil {
		return []*Photo{}, err
	}

	return result, nil
}

func GetFavoritePhotosByUserId(userId int64, offset int, limit int) ([]*Photo, error) {
	result := make([]*Photo, 0, 1)

	rows, err := Db.Query(
		`
		SELECT
			p.id,
			u.id,
			COALESCE(u.username, ''),
			u.realname,
			p.rand_id,
			p.tm,
			p.processed,
			p.title,
			p.description,
			p.views_count,
			(SELECT COUNT(*) FROM comments c WHERE c.photo_id = p.id),
			(SELECT COUNT(*) FROM favorites f WHERE f.photo_id = p.id) 
		FROM 
			photos p 
			JOIN users u ON u.id = p.user_id,
			favorites f
		WHERE 
			f.user_id = $1
			AND f.photo_id = p.id
			AND p.processed = 1
		ORDER BY p.tm DESC
		OFFSET $2
		LIMIT $3
		`,
		userId, offset, limit,
	)

	if err != nil {
		return []*Photo{}, err
	}

	for rows.Next() {
		photo := &Photo{}
		err := rows.Scan(
			&photo.Id,
			&photo.UserId, &photo.UserUsername, &photo.UserRealName,
			&photo.RandId, &photo.Tm, &photo.Processed, &photo.Title,
			&photo.Description, &photo.ViewsCount,
			&photo.CommentsCount, &photo.FavoritesCount,
		)
		if err != nil {
			return []*Photo{}, err
		}
		result = append(result, photo)
	}

	if err := rows.Err(); err != nil {
		return []*Photo{}, err
	}

	return result, nil
}

func GetLatestPhotos(offset int, limit int) ([]*Photo, error) {
	result := make([]*Photo, 0, 1)

	rows, err := Db.Query(
		`
		SELECT
			p.id,
			u.id,
			COALESCE(u.username, ''),
			u.realname,
			p.rand_id,
			p.tm,
			p.processed,
			p.title,
			p.description,
			p.views_count,
			(SELECT COUNT(*) FROM comments c WHERE c.photo_id = p.id),
			(SELECT COUNT(*) FROM favorites f WHERE f.photo_id = p.id) 
		FROM 
			photos p 
			JOIN users u ON u.id=p.user_id
		WHERE 
			p.processed = 1
		ORDER BY p.tm DESC
		OFFSET $1
		LIMIT $2
		`,
		offset, limit,
	)

	if err != nil {
		return []*Photo{}, err
	}

	for rows.Next() {
		photo := &Photo{}
		err := rows.Scan(
			&photo.Id,
			&photo.UserId, &photo.UserUsername, &photo.UserRealName,
			&photo.RandId, &photo.Tm, &photo.Processed, &photo.Title,
			&photo.Description, &photo.ViewsCount,
			&photo.CommentsCount, &photo.FavoritesCount,
		)
		if err != nil {
			return []*Photo{}, err
		}
		result = append(result, photo)
	}

	if err := rows.Err(); err != nil {
		return []*Photo{}, err
	}

	return result, nil
}
