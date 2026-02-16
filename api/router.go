package api

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var mySigningKey = []byte(os.Getenv("SPEEDSALESJWTKEY"))

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

func NewRouter() *mux.Router {
	// rentals.CreateTables()

	r := mux.NewRouter()

	// r.HandleFunc("/ws", socketHandler)

	r.HandleFunc("/login", loginRoutes).Methods("POST", "OPTIONS")

	r.HandleFunc("/users/{module}", UserPOSTRoutes).Methods("POST", "OPTIONS")
	r.HandleFunc("/login/reset", ResetUserRoutes).Methods("POST", "OPTIONS")

	// r.HandleFunc("/sms", sms.Post).Methods("POST", "OPTIONS")

	return r
}
