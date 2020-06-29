package main

import (
	"app/controller"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/register", controller.RegisterHandler).
		Methods("POST")
	r.HandleFunc("/login", controller.LoginHandler).
		Methods("POST")
	r.HandleFunc("/filter", controller.GetPostsHandler).
		Methods("GET")
	r.HandleFunc("/save-posts", controller.SavePostsHandler).
		Methods("POST")
	r.HandleFunc("/create-post", controller.CreatePostHandler).
		Methods("POST")
	r.HandleFunc("/get-post", controller.GetPostHandler).
		Methods("GET")
	r.HandleFunc("/update-post", controller.UpdatePostHandler).
		Methods("PUT")
	r.HandleFunc("/delete-post", controller.DeletePostHandler).
		Methods("DELETE")
	log.Fatal(http.ListenAndServe(":3000", r))
}
