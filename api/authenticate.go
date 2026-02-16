package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_login/internal/authentication"
)

func loginRoutes(w http.ResponseWriter, r *http.Request) {
	respMap := authentication.POST(w, r)

	jstr, _ := json.Marshal(respMap)

	EnableCors(&w)
	w.Write(jstr)
}
