package main

import (
	"github.com/dchest/uniuri"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	DIRECTORY = "/tmp/"
	ADDRESS   = "http://localhost:8080"
	LENGTH    = 4
	TEXT      = "$ <command> | curl -F 'paste=<-'" + ADDRESS + "\n"
	PORT      = ":8080"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func exists(location string) bool {
	if _, err := os.Stat(location); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true

}

func generateName() string {
	s := uniuri.NewLen(LENGTH)
	file := exists(DIRECTORY + s)
	if file == true {
		generateName()
	}

	return s

}
func save(raw []byte) string {
	paste := raw[92 : len(raw)-46]

	s := generateName()
	location := DIRECTORY + s

	err := ioutil.WriteFile(location, paste, 0644)
	check(err)

	return s
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		param := r.URL.RawQuery
		if param != "" {
			d := DIRECTORY + param
			s, err := ioutil.ReadFile(d)
			check(err)
			io.WriteString(w, string(s))
		} else {
			io.WriteString(w, TEXT)
		}
	case "POST":
		buf, err := ioutil.ReadAll(r.Body)
		check(err)
		io.WriteString(w, ADDRESS+"?"+save(buf)+"\n")
	case "DELETE":
		// Remove the record.
	}
}

func main() {
	http.HandleFunc("/", pasteHandler)
	err := http.ListenAndServe(PORT, nil)
	check(err)

}
