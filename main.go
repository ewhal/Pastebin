package main

import (
	"io/ioutil"
	"net/http"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func pasteHandler(w http.ResponseWriter, req *http.Request) {
	buf, _ := ioutil.ReadAll(req.Body)
	paste := buf[89 : len(buf)-46]
	err := ioutil.WriteFile("/tmp/dat1", paste, 0644)
	check(err)
}

func main() {
	http.HandleFunc("/", pasteHandler)
	http.ListenAndServe(":8080", nil)

}
