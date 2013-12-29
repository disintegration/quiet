package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"os"

	"github.com/disintegration/imaging"
)

const PathToPhotos = "static/photos/"
const PathToAvatars = "static/avatars/"

var PhotoProcessingQueue chan *Photo
var AvatarProcessingQueue chan *User

type ResizeMode struct {
	Width  int
	Height int
	Type   string
	Suffix string
}

var PhotoSizes = []ResizeMode{
	{50, 50, "thumbnail", "t50"},
	{100, 100, "thumbnail", "t100"},
	{200, 200, "thumbnail", "t200"},
	{200, 200, "fit", "f200"},
	{300, 300, "fit", "f300"},
	{500, 500, "fit", "f500"},
	{1000, 1000, "fit", "f1000"},
	{2000, 2000, "fit", "f2000"},
}

var AvatarSizes = []ResizeMode{
	{25, 25, "thumbnail", "25"},
	{50, 50, "thumbnail", "50"},
	{75, 75, "thumbnail", "75"},
	{200, 200, "thumbnail", "200"},
}

func EnqueuePhoto(photo *Photo) {
	go func() { PhotoProcessingQueue <- photo }()
}

func EnqueueAvatar(user *User) {
	go func() { AvatarProcessingQueue <- user }()
}

func StartProcessing() {
	PhotoProcessingQueue = make(chan *Photo, 100)
	AvatarProcessingQueue = make(chan *User, 100)

	go workerProcessPhotos()
	go workerProcessAvatars()
}

func workerProcessPhotos() {
	for photo := range PhotoProcessingQueue {
		err := ProcessPhoto(photo)
		if err != nil {
			SetPhotoProcessed(photo.Id, -1)
			log.Printf("ERROR: cannot process photo %d: %v\n", photo.Id, err)
		} else {
			SetPhotoProcessed(photo.Id, 1)
		}
	}
}

func workerProcessAvatars() {
	for user := range AvatarProcessingQueue {
		err := ProcessAvatar(user)
		if err != nil {
			log.Printf("ERROR: cannot process avatar %d: %v\n", user.Id, err)
		}
	}
}

func GetPhotoPath(photo *Photo, suffix string) string {
	return PathToPhotos + fmt.Sprintf("%d_%s_%s.jpg", photo.Id, photo.RandId, suffix)
}

func GetAvatarPath(user *User, suffix string) string {
	return PathToAvatars + fmt.Sprintf("%d_%s.jpg", user.Id, suffix)
}

func ProcessPhoto(photo *Photo) error {
	origImg, err := imaging.Open(GetPhotoPath(photo, "o"))
	if err != nil {
		return err
	}

	src := imaging.Clone(origImg)

	for _, sz := range PhotoSizes {
		var dst image.Image

		switch sz.Type {
		case "thumbnail":
			dst = imaging.Thumbnail(src, sz.Width, sz.Height, imaging.Lanczos)
		case "fit":
			dst = imaging.Fit(src, sz.Width, sz.Height, imaging.Lanczos)
		}

		err := imaging.Save(dst, GetPhotoPath(photo, sz.Suffix))
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessAvatar(user *User) error {
	origImg, err := imaging.Open(GetAvatarPath(user, "o"))
	if err != nil {
		return err
	}

	src := imaging.Clone(origImg)

	for _, sz := range AvatarSizes {
		var dst image.Image

		switch sz.Type {
		case "thumbnail":
			dst = imaging.Thumbnail(src, sz.Width, sz.Height, imaging.Lanczos)
		case "fit":
			dst = imaging.Fit(src, sz.Width, sz.Height, imaging.Lanczos)
		}

		err := imaging.Save(dst, GetAvatarPath(user, sz.Suffix))
		if err != nil {
			return err
		}
	}

	return nil
}

func SetDefaultAvatar(user *User) error {
	for _, sz := range AvatarSizes {
		defaultPath := PathToAvatars + fmt.Sprintf("default_%s.jpg", sz.Suffix)
		newPath := GetAvatarPath(user, sz.Suffix)

		err := CopyFile(defaultPath, newPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()

	dstf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	return err
}
