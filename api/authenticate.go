package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_login/internal/authentication"
)

// loginRoutes
func loginRoutes(w http.ResponseWriter, r *http.Request) {
	respMap := authentication.POST(w, r)

	jstr, _ := json.Marshal(respMap)

	EnableCors(&w)
	w.WriteHeader(http.StatusOK)
	if respMap["message"] == "wrong username or password" {
		w.WriteHeader(http.StatusUnauthorized)
	}
	w.Write(jstr)
}
	