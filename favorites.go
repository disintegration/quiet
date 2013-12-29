package main

import (
	"time"
)

type Favorite struct {
	Id      int64
	UserId  int64
	PhotoId int64
	Tm      time.Time

	UserUsername string
	UserRealName string
}

func CreateFavorite(userId int64, photoId int64) (*Favorite, error) {
	fav := &Favorite{
		UserId:  userId,
		PhotoId: photoId,
		Tm:      time.Now(),
	}

	err := Db.QueryRow(
		`
		INSERT INTO favorites(user_id, photo_id, tm) 
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		fav.UserId, fav.PhotoId, fav.Tm,
	).Scan(&fav.Id)

	if err != nil {
		return nil, err
	}
	return fav, nil
}

func GetFavoritesByPhotoId(photoId int64) ([]*Favorite, error) {
	result := make([]*Favorite, 0, 10)

	rows, err := Db.Query(
		`
		SELECT f.id, u.id, u.username, u.realname, f.photo_id, f.tm
		FROM 
			favorites f
			JOIN users u ON f.user_id = u.id
		WHERE f.photo_id = $1
		ORDER BY tm DESC
		`,
		photoId,
	)

	if err != nil {
		return []*Favorite{}, err
	}

	for rows.Next() {
		fav := &Favorite{}
		err := rows.Scan(
			&fav.Id, &fav.UserId, &fav.UserUsername, &fav.UserRealName, &fav.PhotoId, &fav.Tm,
		)
		if err != nil {
			return []*Favorite{}, err
		}
		result = append(result, fav)
	}

	if err := rows.Err(); err != nil {
		return []*Favorite{}, err
	}

	return result, nil
}

func GetFavoritesCountByPhotoId(photoId int64) (int, error) {
	var count int

	err := Db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE photo_id = $1`, photoId).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetFavoritesCountByUserId(userId int64) (int, error) {
	var count int

	err := Db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE user_id = $1`, userId).Scan(&count)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func IsFavorited(userId int64, photoId int64) (bool, error) {
	var count int

	err := Db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = $1 AND photo_id = $2`,
		userId, photoId,
	).Scan(&count)

	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func DelFavorite(userId int64, photoId int64) error {
	_, err := Db.Exec(`DELETE FROM favorites WHERE user_id = $1 and photo_id = $2`, userId, photoId)
	return err
}
