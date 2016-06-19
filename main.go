package main

import (
	"database/sql"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/ewhal/pygments"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ADDRESS = "https://p.pantsu.cat/"
	LENGTH  = 6
	TEXT    = "$ <command> | curl -F 'p=<-' " + ADDRESS + "\n"
	PORT    = ":9900"
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func generateName() string {
	s := uniuri.NewLen(LENGTH)
	db, err := sql.Open("sqlite3", "./database.db")
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
func save(raw []byte) string {
	paste := raw[86 : len(raw)-46]

	s := generateName()
	db, err := sql.Open("sqlite3", "./database.db")
	check(err)
	stmt, err := db.Prepare("INSERT INTO pastebin(id, data) values(?,?)")
	_, err = stmt.Exec(s, html.EscapeString(string(paste)))
	check(err)
	db.Close()

	return s
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		param1 := html.EscapeString(r.URL.Query().Get("p"))
		param2 := html.EscapeString(r.URL.Query().Get("lang"))
		db, err := sql.Open("sqlite3", "./database.db")
		var s string
		err = db.QueryRow("select data from pastebin where id=?", param1).Scan(&s)
		db.Close()
		check(err)

		if param1 != "" {

			if err == sql.ErrNoRows {
				io.WriteString(w, "Error invalid paste")
			}
			if param2 != "" {
				highlight := pygments.Highlight(html.UnescapeString(s), param2, "html", "full, style=autumn,linenos=True, lineanchors=True,anchorlinenos=True,", "utf-8")
				io.WriteString(w, highlight)

			} else {
				io.WriteString(w, html.UnescapeString(s))
			}
		} else {
			io.WriteString(w, TEXT)
		}
	case "POST":
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		name := save(buf)
		io.WriteString(w, ADDRESS+"?p="+name+"\n")
	case "DELETE":
		// Remove the record.
	}
}

func main() {
	http.HandleFunc("/", pasteHandler)
	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		panic(err)
	}

}
