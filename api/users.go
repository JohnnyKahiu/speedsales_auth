package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_login/internal/user"
)

func UserPOSTRoutes(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("\n\t User POST Routes\n\n")
	respMap := user.UserPOSTRoutes(w, r)

	EnableCors(&w)
	jResp, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jResp)
}

// ResetUserRoutes handles the reset of a user's password routes
// it writes the response to the response writer
func ResetUserRoutes(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n\t reseting user \n\n")
	respMap := user.POSTLoginResetRequest(w, r)

	EnableCors(&w)
	jResp, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jResp)
}
