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

	status := http.StatusOK
	if respMap["message"] == "wrong username or password" {
		status = http.StatusUnauthorized
	}
	w.WriteHeader(status)
	w.Write(jstr)
}
