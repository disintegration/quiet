package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/securecookie"
)

var Sc *securecookie.SecureCookie

func init() {
	Sc = securecookie.New(
		[]byte("6fbbd7p95oe8ut5qrttebiwar88s74do"),
		[]byte("nnl9x8lipmar2rysjl0p0f5p9u8nz7lc"),
	)
}

func SetSecureCookie(w http.ResponseWriter, name, value string) {
	if encoded, err := Sc.Encode(name, value); err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:  name,
			Value: encoded,
			Path:  "/",
		})
	}
}

func GetSecureCookie(r *http.Request, name string) (string, error) {
	if cookie, err := r.Cookie(name); err == nil {
		var value string
		if err = Sc.Decode(name, cookie.Value, &value); err == nil {
			return value, nil
		}
	}
	return "", fmt.Errorf("No such cookie: %s", name)
}

func SetCurrentUser(w http.ResponseWriter, user *User) {
	userIdStr := strconv.FormatInt(user.Id, 10)
	SetSecureCookie(w, "user", userIdStr)
}

func GetCurrentUser(r *http.Request) *User {
	var err error
	var userIdStr string
	var id int64
	var user *User

	userIdStr, err = GetSecureCookie(r, "user")
	if err != nil {
		return nil
	}

	id, err = strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return nil
	}

	user, err = GetUserById(id)
	if err != nil {
		return nil
	}

	return user
}

func ClearCurrentUser(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "user",
		Value:  "",
		Path:   "/",
		MaxAge: 0,
	})
}

func SetLayout(w http.ResponseWriter, layout string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "layout",
		Value: layout,
	})
}

func GetLayout(r *http.Request) string {
	if cookie, err := r.Cookie("layout"); err == nil {
		layout := cookie.Value
		if layout == "S" || layout == "M" || layout == "L" {
			return layout
		}
	}
	return "S"
}

func GetPhotoSuffixByLayout(layout string) string {
	switch layout {
	case "S":
		return "f300"
	case "M":
		return "f500"
	case "L":
		return "f1000"
	}
	return "t50"
}
