package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "gke-tpu-env-injector")
	})
	srv := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}
	log.Println("Listening on :443")
	err := srv.ListenAndServeTLS("/etc/tls/tls.crt", "/etc/tls/tls.key")
	if err != nil {
		panic(err)
	}
}
