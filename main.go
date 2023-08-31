package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "gke-tpu-env-injector")
	})
	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	_ = srv.ListenAndServe()
}
