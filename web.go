package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var Tp *template.Template

func main() {
	var err error

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Database initialization

	DbConnect()
	DbInitSchema()

	// Start photos and avatars processing

	StartProcessing()

	// Prepare templates

	var funcMap = template.FuncMap{
		"formatdt": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
	}

	Tp, err = template.New("Tp").Funcs(funcMap).ParseFiles(
		"templates/header.html",
		"templates/footer.html",
		"templates/home.html",
		"templates/settings.html",
		"templates/upload.html",
		"templates/photostream.html",
		"templates/photo.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup routes

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc(`/`, HandleHome)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/`, HandleUserPhotos)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/page/{page:\d+}/`, HandleUserPhotos)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/`, HandlePhoto)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/fav/`, HandleAddFavorite)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/unfav/`, HandleDeleteFavorite)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/del/`, HandleDeletePhoto)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/comment/`, HandleAddComment)
	r.HandleFunc(`/photos/{username:[a-z0-9_]+}/{photo:\d+}/delcomment/{id:\d+}/`, HandleDeleteComment)
	r.HandleFunc(`/favorites/{username:[a-z0-9_]+}/`, HandleUserFavorites)
	r.HandleFunc(`/favorites/{username:[a-z0-9_]+}/page/{page:\d+}/`, HandleUserFavorites)
	r.HandleFunc(`/contacts/add/{username:[a-z0-9_]+}/`, HandleAddContact)
	r.HandleFunc(`/contacts/del/{username:[a-z0-9_]+}/`, HandleDeleteContact)
	r.HandleFunc(`/contacts/photos/`, HandleContactsPhotos)
	r.HandleFunc(`/contacts/photos/page/{page:\d+}/`, HandleContactsPhotos)
	r.HandleFunc(`/settings/`, HandleSettings)
	r.HandleFunc(`/upload/`, HandleUpload)
	r.HandleFunc(`/login/`, HandleLogin)
	r.HandleFunc(`/logout/`, HandleLogout)

	r.PathPrefix(`/static/`).Handler(http.StripPrefix("/static/", &StaticFileHandler{"static/"}))

	// Start the server

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	http.Handle("/", r)
	http.ListenAndServe(":"+port, nil)
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	var err error
	currentUser := GetCurrentUser(r)

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	var userPhotos []*Photo
	var contactsPhotos []*Photo
	var othersPhotos []*Photo

	if currentUser != nil {
		userPhotos, err = GetPhotosByUserId(currentUser.Id, 0, 5)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		contactsPhotos, err = GetContactsPhotosByUserId(currentUser.Id, 0, 5)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	othersPhotos, err = GetLatestPhotos(0, 15)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = Tp.ExecuteTemplate(w, "home.html",
		struct {
			CurrentUser    *User
			UserPhotos     []*Photo
			ContactsPhotos []*Photo
			OthersPhotos   []*Photo
		}{
			CurrentUser:    currentUser,
			UserPhotos:     userPhotos,
			ContactsPhotos: contactsPhotos,
			OthersPhotos:   othersPhotos,
		},
	)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func HandleUserPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	pageStr := vars["page"]
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	currentUser := GetCurrentUser(r)

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	limit := 30
	offset := (int(page) - 1) * limit
	photos, err := GetPhotosByUserId(user.Id, offset, limit)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	photosCount, err := GetPhotosCountByUserId(user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	lastPage := int64(1 + photosCount/limit)

	layout := GetLayout(r)
	suffix := GetPhotoSuffixByLayout(layout)

	showAddContact := false
	showDelContact := false

	if currentUser != nil && currentUser.Id != user.Id {
		res, err := IsContacted(currentUser.Id, user.Id)
		if err == nil {
			if res {
				showDelContact = true
			} else {
				showAddContact = true
			}
		}
	}

	err = Tp.ExecuteTemplate(w, "photostream.html",
		struct {
			PhotostreamType string
			PhotostreamUrl  string
			CurrentUser     *User
			User            *User
			Photos          []*Photo
			Page            int64
			PrevPage        int64
			NextPage        int64
			LastPage        int64
			Layout          string
			PhotoSuffix     string
			ShowAddContact  bool
			ShowDelContact  bool
			ShowPhotoAuthor bool
		}{
			PhotostreamType: "user-photos",
			PhotostreamUrl:  fmt.Sprintf("/photos/%s/", user.Username),
			CurrentUser:     currentUser,
			User:            user,
			Photos:          photos,
			Page:            page,
			PrevPage:        page - 1,
			NextPage:        page + 1,
			LastPage:        lastPage,
			Layout:          layout,
			PhotoSuffix:     suffix,
			ShowAddContact:  showAddContact,
			ShowDelContact:  showDelContact,
			ShowPhotoAuthor: false,
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func HandlePhoto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	showAddContact := false
	showDelContact := false

	if currentUser != nil && currentUser.Id != user.Id {
		res, err := IsContacted(currentUser.Id, user.Id)
		if err == nil {
			if res {
				showDelContact = true
			} else {
				showAddContact = true
			}
		}
	}

	showAddFavorite := false
	showDelFavorite := false

	if currentUser != nil && currentUser.Id != user.Id {
		res, err := IsFavorited(currentUser.Id, photo.Id)
		if err == nil {
			if res {
				showDelFavorite = true
			} else {
				showAddFavorite = true
			}
		}
	}

	comments, err := GetCommentsByPhotoId(photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if currentUser == nil || currentUser.Id != user.Id {
		IncPhotoViewsCount(photo.Id)
	}

	err = Tp.ExecuteTemplate(w, "photo.html",
		struct {
			CurrentUser     *User
			User            *User
			Photo           *Photo
			ShowAddContact  bool
			ShowDelContact  bool
			ShowAddFavorite bool
			ShowDelFavorite bool
			Comments        []*Comment
		}{
			CurrentUser:     currentUser,
			User:            user,
			Photo:           photo,
			ShowAddContact:  showAddContact,
			ShowDelContact:  showDelContact,
			ShowAddFavorite: showAddFavorite,
			ShowDelFavorite: showDelFavorite,
			Comments:        comments,
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if currentUser.Id == photo.UserId {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	res, err := IsFavorited(currentUser.Id, photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if res {
		fmt.Fprintln(w, "OK")
		return
	}

	_, err = CreateFavorite(currentUser.Id, photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleDeleteFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if currentUser.Id == photo.UserId {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	res, err := IsFavorited(currentUser.Id, photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if !res {
		fmt.Fprintln(w, "OK")
		return
	}

	err = DelFavorite(currentUser.Id, photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if photo.UserId != currentUser.Id {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// TODO: Remove images from disk first
	err = DelPhotoById(photo.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleAddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	commentText := r.FormValue("comment")
	_, err = CreateComment(currentUser.Id, photo.Id, commentText)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	photoIdStr := vars["photo"]
	photoId, err := strconv.ParseInt(photoIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	commentIdStr := vars["id"]
	commentId, err := strconv.ParseInt(commentIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	photo, err := GetPhotoById(photoId)
	if err != nil || photo.UserId != user.Id {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	comment, err := GetCommentById(commentId)
	if err != nil || comment.PhotoId != photo.Id {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if currentUser.Id != user.Id && currentUser.Id != comment.UserId {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	err = DelCommentById(commentId)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleUserFavorites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	pageStr := vars["page"]
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	currentUser := GetCurrentUser(r)

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	limit := 30
	offset := (int(page) - 1) * limit
	photos, err := GetFavoritePhotosByUserId(user.Id, offset, limit)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	photosCount, err := GetFavoritesCountByUserId(user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	lastPage := int64(1 + photosCount/limit)

	layout := GetLayout(r)
	suffix := GetPhotoSuffixByLayout(layout)

	showAddContact := false
	showDelContact := false

	if currentUser != nil && currentUser.Id != user.Id {
		res, err := IsContacted(currentUser.Id, user.Id)
		if err == nil {
			if res {
				showDelContact = true
			} else {
				showAddContact = true
			}
		}
	}

	err = Tp.ExecuteTemplate(w, "photostream.html",
		struct {
			PhotostreamType string
			PhotostreamUrl  string
			CurrentUser     *User
			User            *User
			Photos          []*Photo
			Page            int64
			PrevPage        int64
			NextPage        int64
			LastPage        int64
			Layout          string
			PhotoSuffix     string
			ShowAddContact  bool
			ShowDelContact  bool
			ShowPhotoAuthor bool
		}{
			PhotostreamType: "user-favorites",
			PhotostreamUrl:  fmt.Sprintf("/favorites/%s/", user.Username),
			CurrentUser:     currentUser,
			User:            user,
			Photos:          photos,
			Page:            page,
			PrevPage:        page - 1,
			NextPage:        page + 1,
			LastPage:        lastPage,
			Layout:          layout,
			PhotoSuffix:     suffix,
			ShowAddContact:  showAddContact,
			ShowDelContact:  showDelContact,
			ShowPhotoAuthor: true,
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func HandleContactsPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageStr := vars["page"]
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	limit := 30
	offset := (int(page) - 1) * limit
	photos, err := GetContactsPhotosByUserId(currentUser.Id, offset, limit)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	photosCount, err := GetContactsPhotosCountByUserId(currentUser.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	lastPage := int64(1 + photosCount/limit)

	layout := GetLayout(r)
	suffix := GetPhotoSuffixByLayout(layout)

	err = Tp.ExecuteTemplate(w, "photostream.html",
		struct {
			PhotostreamType string
			PhotostreamUrl  string
			CurrentUser     *User
			User            *User
			Photos          []*Photo
			Page            int64
			PrevPage        int64
			NextPage        int64
			LastPage        int64
			Layout          string
			PhotoSuffix     string
			ShowAddContact  bool
			ShowDelContact  bool
			ShowPhotoAuthor bool
		}{
			PhotostreamType: "contacts-photos",
			PhotostreamUrl:  "/contacts/photos/",
			CurrentUser:     currentUser,
			User:            currentUser,
			Photos:          photos,
			Page:            page,
			PrevPage:        page - 1,
			NextPage:        page + 1,
			LastPage:        lastPage,
			Layout:          layout,
			PhotoSuffix:     suffix,
			ShowAddContact:  false,
			ShowDelContact:  false,
			ShowPhotoAuthor: true,
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func HandleAddContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if currentUser.Id == user.Id {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	res, err := IsContacted(currentUser.Id, user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if res {
		fmt.Fprintln(w, "OK")
		return
	}

	_, err = CreateContact(currentUser.Id, user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if currentUser.Id == user.Id {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	res, err := IsContacted(currentUser.Id, user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if !res {
		fmt.Fprintln(w, "OK")
		return
	}

	err = DelContact(currentUser.Id, user.Id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "OK")
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if currentUser != nil && currentUser.Username == "" {
		http.Redirect(w, r, "/settings/", http.StatusFound)
		return
	}

	switch r.Method {

	case "GET":

		err := Tp.ExecuteTemplate(w, "upload.html",
			struct {
				CurrentUser *User
			}{
				CurrentUser: currentUser,
			},
		)

		if err != nil {
			log.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

	case "POST":

		pTitle := r.FormValue("photo_title")
		pDesc := r.FormValue("photo_description")

		photoFile, _, err := r.FormFile("photo_file")
		if err != nil {
			http.Error(w, "Upload error", http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(photoFile)
		if err != nil {
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}

		photo, err := CreatePhoto(currentUser.Id, pTitle, pDesc)
		if err != nil {
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}

		origPhotoPath := GetPhotoPath(photo, "o")
		err = ioutil.WriteFile(origPhotoPath, data, 0777)
		if err != nil {
			DelPhotoById(photo.Id)
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}

		EnqueuePhoto(photo)

		userPhotoPage := fmt.Sprintf("/photos/%s/", currentUser.Username)
		http.Redirect(w, r, userPhotoPage, http.StatusFound)

	default:

		http.Error(w, "Bad request", http.StatusBadRequest)

	}
}

func HandleSettings(w http.ResponseWriter, r *http.Request) {
	currentUser := GetCurrentUser(r)

	if currentUser == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	switch r.Method {

	case "GET":

		err := Tp.ExecuteTemplate(w, "settings.html",
			struct {
				CurrentUser *User
			}{
				CurrentUser: currentUser,
			},
		)

		if err != nil {
			log.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

	case "POST":

		username := r.FormValue("username")
		if currentUser.Username == "" && username != "" {
			matched, _ := regexp.MatchString("^[a-z0-9_]{3,30}$", username)
			if matched {
				_, err := GetUserByUsername(username)
				if err.Error() == "User not found" {
					currentUser.Username = username
				}
			}
		}

		realName := r.FormValue("realname")
		currentUser.RealName = realName

		UpdateUser(currentUser)

		avFile, _, err := r.FormFile("avatar_file")
		if err == nil {
			data, err := ioutil.ReadAll(avFile)
			if err == nil {
				origAvPath := GetAvatarPath(currentUser, "o")
				err = ioutil.WriteFile(origAvPath, data, 0777)
				if err == nil {
					EnqueueAvatar(currentUser)
				}
			}
		}

		http.Redirect(w, r, "/settings/", http.StatusFound)

	default:

		http.Error(w, "Bad request", http.StatusBadRequest)

	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	audience := r.Host
	assertion := r.FormValue("assertion")
	if assertion == "" {
		http.Error(w, "Authentication error", http.StatusBadRequest)
		return
	}

	values := url.Values{"audience": {audience}, "assertion": {assertion}}

	resp, err := http.PostForm("https://verifier.login.persona.org/verify", values)
	if err != nil {
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Status   string
		Email    string
		Audience string
		Issuer   string
		Expires  int64
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil || data.Status != "okay" || data.Email == "" {
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	user, err := GetUserByPersona(data.Email)
	if err != nil {
		user, err = CreatePersonaUser(data.Email)
		if err != nil {
			http.Error(w, "Authentication error", http.StatusInternalServerError)
			return
		}

		err = SetDefaultAvatar(user)
		if err != nil {
			http.Error(w, "Authentication error", http.StatusInternalServerError)
			return
		}
	}

	SetCurrentUser(w, user)
	fmt.Fprintln(w, "OK")
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ClearCurrentUser(w)
	fmt.Fprintln(w, "OK")
}

type StaticFileHandler struct {
	StaticDir string
}

func (h *StaticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "HEAD" {
		pth := filepath.Join(h.StaticDir, r.URL.Path)

		if stat, err := os.Stat(pth); err == nil {
			if stat.Mode().IsRegular() {
				http.ServeFile(w, r, pth)
				return
			}
		}
	}

	http.NotFound(w, r)
	return
}
