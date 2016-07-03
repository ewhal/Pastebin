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
	TITLE  string `json:"title"`
	HASH   string `json:"hash"`
	URL    string `json:"url"`
	SIZE   int    `json:"size"`
	DELKEY string `json:"delkey"`
}

type Page struct {
	Title    string
	Body     []byte
	Raw      string
	Home     string
	Download string
	Clone    string
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func generateName() string {
	id := uniuri.NewLen(LENGTH)
	db, err := sql.Open("mysql", DATABASE)
	check(err)

	query, err := db.Query("select id from pastebin where id=?", id)
	if err != sql.ErrNoRows {
		for query.Next() {
			generateName()
		}
	}
	db.Close()

	return id

}
func hash(paste string) string {
	hasher := sha1.New()

	hasher.Write([]byte(paste))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
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

func save(raw string, lang string, title string, expiry string) []string {
	db, err := sql.Open("mysql", DATABASE)
	check(err)

	sha := hash(raw)
	query, err := db.Query("select id, title, hash, data, delkey from pastebin where hash=?", sha)
	if err != sql.ErrNoRows {
		for query.Next() {
			var id, title, hash, paste, delkey string
			err := query.Scan(&id, &title, &hash, &paste, &delkey)
			check(err)
			url := ADDRESS + "/p/" + id
			return []string{id, title, hash, url, paste, delkey}
		}
	}
	id := generateName()
	var url string
	if lang == "" {
		url = ADDRESS + "/p/" + id
	} else {
		url = ADDRESS + "/p/" + id + "/" + lang
	}
	const timeFormat = "2006-01-02 15:04:05"
	expiryTime := time.Now().Add(durationFromExpiry(expiry)).Format(timeFormat)

	delKey := uniuri.NewLen(40)
	dataEscaped := html.EscapeString(raw)

	stmt, err := db.Prepare("INSERT INTO pastebin(id, title, hash, data, delkey, expiry) values(?,?,?,?,?,?)")
	check(err)
	if title == "" {
		_, err = stmt.Exec(id, id, sha, dataEscaped, delKey, expiryTime)
		check(err)
	} else {
		_, err = stmt.Exec(id, html.EscapeString(title), sha, dataEscaped, delKey, expiryTime)
		check(err)
	}
	db.Close()
	return []string{id, title, sha, url, dataEscaped, delKey}
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["pasteId"]
	delkey := vars["delKey"]

	db, err := sql.Open("mysql", DATABASE)
	check(err)

	stmt, err := db.Prepare("delete from pastebin where delkey=? and id=?")
	check(err)

	res, err := stmt.Exec(html.EscapeString(delkey), html.EscapeString(id))
	check(err)

	_, err = res.RowsAffected()
	if err == sql.ErrNoRows {
		io.WriteString(w, "Error invalid paste")
	} else {
		io.WriteString(w, id+" deleted")
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
		title := r.FormValue("title")
		expiry := r.FormValue("expiry")
		if paste == "" {
			http.Error(w, "Empty paste", 500)
			return
		}
		values := save(paste, lang, title, expiry)
		b := &Response{
			ID:     values[0],
			TITLE:  values[1],
			HASH:   values[2],
			URL:    values[3],
			SIZE:   len(values[4]),
			DELKEY: values[5],
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

func highlight(s string, lang string) (string, error) {

	highlight, err := pygments.Highlight(html.UnescapeString(s), html.EscapeString(lang), "html", "style=autumn,linenos=True, lineanchors=True,anchorlinenos=True,noclasses=True,", "utf-8")
	if err != nil {
		return "", err
	}
	return highlight, nil

}

func getPaste(paste string, lang string) (string, string) {
	param1 := html.EscapeString(paste)
	db, err := sql.Open("mysql", DATABASE)
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
	db.Close()

	if err == sql.ErrNoRows {
		return "Error invalid paste", ""
	} else {
		if lang == "" {
			return html.UnescapeString(s), html.UnescapeString(title)
		} else {
			high, err := highlight(s, lang)
			check(err)
			return high, html.UnescapeString(title)

		}
	}

}

var templates = template.Must(template.ParseFiles("assets/paste.html", "assets/index.html", "assets/clone.html"))
var syntax, _ = ioutil.ReadFile("assets/syntax.html")

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	lang := vars["lang"]
	s, title := getPaste(paste, lang)
	link := ADDRESS + "/raw/" + paste
	download := ADDRESS + "/download/" + paste
	clone := ADDRESS + "/clone/" + paste
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
	link := ADDRESS + "/raw/" + paste
	download := ADDRESS + "/download/" + paste
	clone := ADDRESS + "/clone/" + paste
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
	w.Header().Set("Content-Disposition", "attachment; filename="+paste)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	io.WriteString(w, s)

}
func rawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paste := vars["pasteId"]
	s, _ := getPaste(paste, "")
	io.WriteString(w, s)

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/p/{pasteId}", pasteHandler)
	router.HandleFunc("/raw/{pasteId}", rawHandler)
	router.HandleFunc("/p/{pasteId}/{lang}", pasteHandler)
	router.HandleFunc("/clone/{pasteId}", cloneHandler)
	router.HandleFunc("/download/{pasteId}", downloadHandler)
	router.HandleFunc("/save", saveHandler)
	router.HandleFunc("/save/{output}", saveHandler)
	router.HandleFunc("/del/{pasteId}/{delKey}", delHandler)
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("assets/"))))
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
