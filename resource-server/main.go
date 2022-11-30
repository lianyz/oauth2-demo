/*
@Time : 2022/11/30 20:36
@Author : lianyz
@Description :
*/

package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, I'm protected!\n"))
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}
