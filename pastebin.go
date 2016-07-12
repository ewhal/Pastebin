// Package pastebin is a simple modern and powerful pastebin service
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
	"io/ioutil"
	"log"
	"net/http"
	"time"

	// uniuri is used for easy random string generation
	"github.com/dchest/uniuri"
	// pygments is used for syntax highlighting
	"github.com/ewhal/pygments"
	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	// mux is used for url routing
	"github.com/gorilla/mux"
)

const (
	// ADDRESS that pastebin will return links for
	ADDRESS = "http://localhost:9900"
	// LENGTH of paste id
	LENGTH = 6
	// PORT that pastebin will listen on
	PORT = ":9900"
	// USERNAME for database
	USERNAME = ""
	// PASS database password
	PASS = ""
	// NAME database name
	NAME = ""
	// DATABASE connection String
	DATABASE = USERNAME + ":" + PASS + "@/" + NAME + "?charset=utf8"
)

// Template pages
var templates = template.Must(template.ParseFiles("assets/paste.html", "assets/index.html", "assets/clone.html"))
var syntax, _ = ioutil.ReadFile("assets/syntax.html")

// Response API struct
type Response struct {
	ID     string `json:"id"`
	TITLE  string `json:"title"`
	HASH   string `json:"hash"`
	URL    string `json:"url"`
	SIZE   int    `json:"size"`
	DELKEY string `json:"delkey"`
}

// Page generation struct
type Page struct {
	Title    string
	Body     []byte
	Raw      string
	Home     string
	Download string
	Clone    string
}

// check error handling function
func check(err error) {
	if err != nil {
		log.Println(err)
	}
}

// generateName uses uniuri to generate a random string that isn't in the
// database
func generateName() string {
	// use uniuri to generate random string
	id := uniuri.NewLen(LENGTH)

	db, err := sql.Open("mysql", DATABASE)
	check(err)
	defer db.Close()
	// query database if id exists and if it does call generateName again
	query, err := db.Query("select id from pastebin where id=?", id)
	if err != sql.ErrNoRows {
		for query.Next() {
			generateName()
		}
	}

	return id

}

// hash hashes paste into a sha1 hash
func hash(paste string) string {
	hasher := sha1.New()

	hasher.Write([]byte(paste))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// durationFromExpiry takes the expiry in string format and returns the duration
// that the paste will exist for
func durationFromExpiry(expiry string) time.Duration {
	switch expiry {
	case "5 minutes":
		return time.Minute * 5
	case "1 hour":
		return time.Hour + 1 // XXX: did you mean '*'?
	case "1 day":
		return time.Hour * 24
	case "1 week":
		return time.Hour * 24 * 7
	case "1 month":
		return time.Hour * 24 * 30
	case "1 year":
		return time.Hour * 24 * 365
	case "forever":
		return time.Hour * 24 * (365 * 20)
	}
	return time.Hour * 24 * (365 * 20)
}

// save function handles the saving of each paste.
// raw string is the raw paste input
// lang string is the user specified language for syntax highlighting
// title string user customized title
// expiry string duration that the paste will exist for
// Returns Response struct
func save(raw string, lang string, title string, expiry string) Response {

	db, err := sql.Open("mysql", DATABASE)
	check(err)
	defer db.Close()

	// hash paste data and query database to see if paste exists
	sha := hash(raw)
	query, err := db.Query("select id, title, hash, data, delkey from pastebin where hash=?", sha)

	if err != sql.ErrNoRows {
		for query.Next() {
			var id, title, hash, paste, delkey string
			err := query.Scan(&id, &title, &hash, &paste, &delkey)
			check(err)
			url := ADDRESS + "/p/" + id
			return Response{id, title, hash, url, len(paste), delkey}
		}
	}
	id := generateName()
	url := ADDRESS + "/p/" + id
	if lang != "" {
		url += "/" + lang
	}

	const timeFormat = "2006-01-02 15:04:05"
	expiryTime := time.Now().Add(durationFromExpiry(expiry)).Format(timeFormat)

	delKey := uniuri.NewLen(40)
	dataEscaped := html.EscapeString(raw)

	stmt, err := db.Prepare("INSERT INTO pastebin(id, title, hash, data, delkey, expiry) values(?,?,?,?,?,?)")
	check(err)
	if title == "" {
		title = id
	}
	_, err = stmt.Exec(id, html.EscapeString(title), sha, dataEscaped, delKey, expiryTime)
	check(err)

	return Response{id, title, sha, url, len(dataEscaped), delKey}
}

// delHandler checks to see if delkey and pasteid exist in the database.
// if both exist and are correct the paste will be removed.
func delHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["pasteId"]
	delkey := vars["delKey"]

	db, err := sql.Open("mysql", DATABASE)
	check(err)
	defer db.Close()

	stmt, err := db.Prepare("delete from pastebin where delkey=? and id=?")
	check(err)

	res, err := stmt.Exec(html.EscapeString(delkey), html.EscapeString(id))
	check(err)

	_, err = res.RowsAffected()
	if err != sql.ErrNoRows {
		io.WriteString(w, id+" deleted")
	}
}

// saveHandler
func saveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	output := vars["output"]
	switch r.Method {
	case "POST":
		paste := r.FormValue("p")
		lang := r.FormValue("lang")
		title := r.FormValue("title")
		expiry := r.FormValue("expiry")
		if paste == "" {
			http.Error(w, "Empty paste", 500)
			return
		}
		b := save(paste, lang, title, expiry)

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
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8; imeanit=yes")
			io.WriteString(w, b.URL+"\n")
			io.WriteString(w, "delete key: "+b.DELKEY+"\n")
		}
	}

}

// highlight uses user specified input to call pygments library to highlight the
// paste
func highlight(s string, lang string) (string, error) {

	highlight, err := pygments.Highlight(html.UnescapeString(s), html.EscapeString(lang), "html", "style=autumn,linenos=True, lineanchors=True,anchorlinenos=True,noclasses=True,", "utf-8")
	if err != nil {
		return "", err
	}
	return highlight, nil

}

// getPaste takes pasteid and language
// queries the database and returns paste data
func getPaste(paste string, lang string) (string, string) {
	param1 := html.EscapeString(paste)
	db, err := sql.Open("mysql", DATABASE)
	check(err)
	defer db.Close()
	var title, s string
	var expiry string
	err = db.QueryRow("select title, data, expiry from pastebin where id=?", param1).Scan(&title, &s, &expiry)
	check(err)
	if time.Now().Format("2006-01-02 15:04:05") > expiry {
		stmt, err := db.Prepare("delete from pastebin where id=?")
		check(err)
		_, err = stmt.Exec(param1)
		check(err)
		return "Error invalid paste", ""
	}

	if err == sql.ErrNoRows {
		return "Error invalid paste", ""
	}
	if lang != "" {
		high, err := highlight(s, lang)
		check(err)
		return high, html.UnescapeString(title)
	}
	return html.UnescapeString(s), html.UnescapeString(title)
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	lang := vars["lang"]

	s, title := getPaste(paste, lang)

	// button links
	link := ADDRESS + "/raw/" + paste
	download := ADDRESS + "/download/" + paste
	clone := ADDRESS + "/clone/" + paste
	// Page struct
	p := &Page{
		Title:    title,
		Body:     []byte(s),
		Raw:      link,
		Home:     ADDRESS,
		Download: download,
		Clone:    clone,
	}
	if lang == "" {

		err := templates.ExecuteTemplate(w, "paste.html", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		fmt.Fprintf(w, string(syntax), p.Title, p.Title, s, p.Home, p.Download, p.Raw, p.Clone)

	}
}

func cloneHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]

	s, title := getPaste(paste, "")

	// Page links
	link := ADDRESS + "/raw/" + paste
	download := ADDRESS + "/download/" + paste
	clone := ADDRESS + "/clone/" + paste

	// Clone page struct
	p := &Page{
		Title:    title,
		Body:     []byte(s),
		Raw:      link,
		Home:     ADDRESS,
		Download: download,
		Clone:    clone,
	}
	err := templates.ExecuteTemplate(w, "clone.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	s, _ := getPaste(paste, "")

	// Set header to an attachment so browser will automatically download it
	w.Header().Set("Content-Disposition", "attachment; filename="+paste)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	io.WriteString(w, s)

}
func rawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	s, _ := getPaste(paste, "")

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8; imeanit=yes")
	// simply write string to browser
	io.WriteString(w, s)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", &Page{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/p/{pasteId}", pasteHandler).Methods("GET")
	router.HandleFunc("/raw/{pasteId}", rawHandler).Methods("GET")
	router.HandleFunc("/p/{pasteId}/{lang}", pasteHandler).Methods("GET")
	router.HandleFunc("/clone/{pasteId}", cloneHandler).Methods("GET")
	router.HandleFunc("/download/{pasteId}", downloadHandler).Methods("GET")
	router.HandleFunc("/p/{output}", saveHandler).Methods("POST")
	router.HandleFunc("/p/{pasteId}/{delKey}", delHandler).Methods("DELETE")
	router.HandleFunc("/", rootHandler)
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
