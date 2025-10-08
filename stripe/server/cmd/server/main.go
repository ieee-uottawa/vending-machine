package main

import (
	"ieeeuottawa/vend-server/internal/router"
	"log"
	"net/http"
)

func main() {
	r := router.NewRouter()

	log.Println("Server starting on port :3000")
	http.ListenAndServe(":3000", r)
}
