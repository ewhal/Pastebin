package main

import (
	"fmt"
	"github.com/dchest/uniuri"
	"io/ioutil"
	"net/http"
	"os"
)

var directory = "/tmp/"
var address = "http://localhost:8080/p/"

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
	s := uniuri.NewLen(4)
	file := exists(directory + s)
	if file == true {
		generateName()
	}

	return s

}
func save(buf []byte) string {
	paste := buf[92 : len(buf)-46]

	s := generateName()
	loc := directory + s

	err := ioutil.WriteFile(loc, paste, 0644)
	check(err)

	url := address + s
	return url
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		text := "$ <command> | curl -F 'paste=<-' " + address
		fmt.Fprintf(w, text)
	case "POST":
		buf, _ := ioutil.ReadAll(r.Body)
		fmt.Fprintf(w, save(buf))
	case "DELETE":
		// Remove the record.
	}
}

func main() {
	http.HandleFunc("/", pasteHandler)
	http.Handle("/p/", http.StripPrefix("/p/", http.FileServer(http.Dir("/tmp"))))

	http.ListenAndServe(":8080", nil)

}
