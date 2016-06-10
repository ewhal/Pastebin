package main

import (
	"fmt"
	"github.com/dchest/uniuri"
	"io/ioutil"
	"net/http"
	"os"
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
	s := uniuri.NewLen(4)
	return s

}
func save(buf []byte) string {
	paste := buf[92 : len(buf)-46]
	address := "localhost:8080/p/"

	dir := "/tmp/"
	s := generateName()
	loc := dir + s

	err := ioutil.WriteFile(loc, paste, 0644)
	check(err)

	url := address + s
	return url
}

func pasteHandler(w http.ResponseWriter, req *http.Request) {
	buf, _ := ioutil.ReadAll(req.Body)
	fmt.Fprintf(w, save(buf))
}

func main() {
	http.HandleFunc("/", pasteHandler)
	http.Handle("/p/", http.StripPrefix("/p/", http.FileServer(http.Dir("/tmp"))))

	http.ListenAndServe(":8080", nil)

}
