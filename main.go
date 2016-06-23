package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/ewhal/pygments"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	ADDRESS  = "http://localhost:9900"
	LENGTH   = 6
	PORT     = ":9900"
	USERNAME = ""
	PASS     = ""
	NAME     = ""
	DATABASE = USERNAME + ":" + PASS + "@/" + NAME + "?charset=utf8"
)

type Response struct {
	ID     string `json:"id"`
	HASH   string `json:"hash"`
	URL    string `json:"url"`
	SIZE   int    `json:"size"`
	DELKEY string `json:"delkey"`
}

type Page struct {
	Title string
	Body  []byte
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func generateName() string {
	s := uniuri.NewLen(LENGTH)
	db, err := sql.Open("mysql", DATABASE)
	check(err)

	query, err := db.Query("select id from pastebin")
	for query.Next() {
		var id string
		err := query.Scan(&id)
		if err != nil {

		}
		if id == s {
			generateName()
		}
	}
	db.Close()

	return s

}
func hash(paste string) string {
	hasher := sha1.New()

	hasher.Write([]byte(paste))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func save(raw string, lang string) []string {
	db, err := sql.Open("mysql", DATABASE)
	check(err)

	sha := hash(raw)
	query, err := db.Query("select id, hash, data, delkey from pastebin")
	for query.Next() {
		var id, hash, paste, delkey string
		err := query.Scan(&id, &hash, &paste, &delkey)
		check(err)
		if hash == sha {
			url := ADDRESS + "/p/" + id
			return []string{id, hash, url, paste, delkey}
		}
	}
	id := generateName()
	var url string
	if lang == "" {
		url = ADDRESS + "/p/" + id
	} else {
		url = ADDRESS + "/p/" + id + "/" + lang
	}
	delKey := uniuri.NewLen(40)
	paste := html.EscapeString(raw)

	stmt, err := db.Prepare("INSERT INTO pastebin(id, hash, data, delkey) values(?,?,?,?)")
	check(err)
	_, err = stmt.Exec(id, sha, paste, delKey)
	check(err)
	db.Close()
	return []string{id, sha, url, paste, delKey}
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	delkey := vars["delKey"]

	db, err := sql.Open("mysql", DATABASE)
	check(err)

	stmt, err := db.Prepare("delete from pastebin where delkey=?")
	check(err)

	res, err := stmt.Exec(html.EscapeString(delkey))
	check(err)

	_, err = res.RowsAffected()
	if err == sql.ErrNoRows {
		io.WriteString(w, "Error invalid paste")
	} else {
		io.WriteString(w, paste+" deleted")
	}
	db.Close()

}
func saveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	output := vars["output"]
	switch r.Method {
	case "POST":
		paste := r.FormValue("p")
		lang := r.FormValue("lang")
		if paste == "" {
			http.Error(w, "Empty paste", 500)
			return
		}
		values := save(paste, lang)
		b := &Response{
			ID:     values[0],
			HASH:   values[1],
			URL:    values[2],
			SIZE:   len(values[3]),
			DELKEY: values[4],
		}

		switch output {
		case "json":

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(b)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "xml":
			x, err := xml.MarshalIndent(b, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/xml")
			w.Write(x)

		case "html":
			w.Header().Set("Content-Type", "text/html")
			io.WriteString(w, "<p><b>URL</b>: <a href='"+b.URL+"'>"+b.URL+"</a></p>")
			io.WriteString(w, "<p><b>Delete Key</b>: <a href='"+ADDRESS+"/del/"+b.ID+"/"+b.DELKEY+"'>"+b.DELKEY+"</a></p>")

		case "redirect":
			http.Redirect(w, r, b.URL, 301)

		default:
			w.Header().Set("Content-Type", "plain/text")
			io.WriteString(w, b.URL+"\n")
			io.WriteString(w, "delete key: "+b.DELKEY+"\n")
		}
	}

}

func langHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	paste := vars["pasteId"]
	lang := vars["lang"]
	s := getPaste(paste)
	highlight, err := pygments.Highlight(html.UnescapeString(s), html.EscapeString(lang), "html", "full, style=autumn,linenos=True, lineanchors=True,anchorlinenos=True,", "utf-8")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, highlight)

}

func getPaste(paste string) string {
	param1 := html.EscapeString(paste)
	db, err := sql.Open("mysql", DATABASE)
	var s string
	err = db.QueryRow("select data from pastebin where id=?", param1).Scan(&s)
	db.Close()
	check(err)

	if err == sql.ErrNoRows {
		return "Error invalid paste"
	} else {
		return html.UnescapeString(s)
	}

}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	s := getPaste(paste)
	p := &Page{
		Title: paste,
		Body:  []byte(s),
	}
	t, err := template.ParseFiles("assets/paste.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, p)
}

func rawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	s := getPaste(paste)
	io.WriteString(w, s)

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/p/{pasteId}", pasteHandler)
	router.HandleFunc("/raw/{pasteId}", rawHandler)
	router.HandleFunc("/p/{pasteId}/{lang}", langHandler)
	router.HandleFunc("/save", saveHandler)
	router.HandleFunc("/save/{output}", saveHandler)
	router.HandleFunc("/del/{pasteId}/{delKey}", delHandler)
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("assets/"))))
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
