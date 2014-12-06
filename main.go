package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB
var codec Codec

func init() {
	codec = newBase64Codec()

	var err error
	db, err = sql.Open("postgres", "database=shortener sslmode=disable")
	if err != nil {
		panic(err)
	}
}

type Codec interface {
	Encode(string) string
	Decode(string) (string, error)
}

type Base64Codec struct {
	e *base64.Encoding
}

func (b Base64Codec) Encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.Replace(str, "=", "", -1)
}

func (b Base64Codec) Decode(s string) (string, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	str, err := base64.URLEncoding.DecodeString(s)
	return string(str), err
}

func newBase64Codec() Base64Codec {
	return Base64Codec{base64.URLEncoding}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleCreate).Methods("POST")
	r.HandleFunc("/{id}", handleFind).Methods("GET")
	http.ListenAndServe(":3000", r)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")

	if len(url) == 0 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "missing url param")
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO urls (url) VALUES ($1) returning id", url).Scan(&id)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(201)
	fmt.Fprintf(w, codec.Encode(strconv.Itoa(id)))
}

func handleFind(w http.ResponseWriter, r *http.Request) {
	url, err := getRecord(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "not found")
		return
	}
	http.Redirect(w, r, url, 301)
}

func getRecord(id string) (url string, err error) {
	rId, err := codec.Decode(id)
	if err != nil {
		return "", err
	}
	err = db.QueryRow("SELECT url from urls WHERE id = $1", rId).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}
