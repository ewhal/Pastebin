package main

import (
	"github.com/dchest/uniuri"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	directory = "/tmp/"
	address   = "http://localhost:8080/p/"
	length    = 4
	text      = "$ <command> | curl -F 'paste=<-'" + address + "\n"
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
	s := uniuri.NewLen(length)
	file := exists(directory + s)
	if file == true {
		generateName()
	}

	return s

}
func save(buf []byte) string {
	paste := buf[92 : len(buf)-46]

	s := generateName()
	location := directory + s

	err := ioutil.WriteFile(location, paste, 0644)
	check(err)

	return s
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		io.WriteString(w, text)
	case "POST":
		buf, _ := ioutil.ReadAll(r.Body)
		io.WriteString(w, address+save(buf)+"\n")
	case "DELETE":
		// Remove the record.
	}
}

func main() {
	http.HandleFunc("/", pasteHandler)
	http.Handle("/p/", http.StripPrefix("/p/", http.FileServer(http.Dir(directory))))

	http.ListenAndServe(":8080", nil)

}
