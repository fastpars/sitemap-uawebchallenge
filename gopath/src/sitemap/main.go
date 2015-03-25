package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello UA Web Challenge 2015")
	})

	fmt.Println("Start listen")
	http.ListenAndServe(":8888", nil)
}
