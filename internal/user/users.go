package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var config = &users.PasswordConfig{
	Time:    1,
	Memory:  64 * 1024,
	Threads: 4,
	KeyLen:  32,
}

func UserPOSTRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	token := r.Header.Get("token")

	details, authentic := users.ValidateToken(r.Context(), token)
	if !authentic {
		respMap["response"] = "error"
		respMap["message"] = "unauthorized"
		return respMap
	}
	if !details.CreateUsers {
		respMap["response"] = "error"
		respMap["message"] = "unauthorized"
		return respMap
	}

	fmt.Println("user_post_routes ")
	fmt.Println("Create_Users =", details.CreateUsers)

	vars := mux.Vars(r)
	m := vars["module"]
	switch m {
	case "create":
		fmt.Println("\t create user")
		var args users.Users
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &args)

		hash, err := users.GeneratePassword(config, "password")
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to generate new password"

			log.Println("error failed to generate new password with     error =", err)
			return respMap
		}

		err = args.CreateUser(hash)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to create user"

			log.Println("error failed to create user with     error =", err)
			return respMap
		}

		respMap["response"] = "success"
		return respMap

	case "update":
		var args map[string]interface{}
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &args)

		fmt.Println("updating username =", string(b))

		err := users.UpdateUser(details, args)
		if err != nil {
			log.Println("error failed to run UpdateUser() with \t err =", err.Error())
			respMap["response"] = "error"
			respMap["message"] = "failed to update user"

			return respMap
		}

		respMap["response"] = "success"
		return respMap

	}

	respMap["response"] = "success"
	return respMap
}

func POSTLoginResetRequest(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	respMap := make(map[string]interface{})

	token := r.Header.Get("token")
	if token == "" {
		respMap["response"] = "error"
		respMap["message"] = "unauthorized"
		return respMap
	}

	user, valid := ValidateToken(ctx, token)
	if !valid {
		respMap["response"] = "error"
		respMap["message"] = "unauthorized"
		return respMap
	}
	// fmt.Println("token =", token)

	// get post body items
	b, err := io.ReadAll(r.Body)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "invalid params"
		return respMap
	}

	// Unmarshal into login_info map
	var loginInfo map[string]string
	err = json.Unmarshal(b, &loginInfo)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "invalid params"
		return respMap
	}

	// compare password 1 and password 2
	if loginInfo["password1"] != loginInfo["password2"] {
		respMap["response"] = "error"
		respMap["message"] = "password mismatch"
		return respMap
	}

	// generate a hash password
	hash, err := users.GeneratePassword(config, loginInfo["password1"])
	fmt.Printf("\n hash = %v", hash)

	// reset password on database
	err = users.ResetPassword(hash, user.Username)
	if err != nil {
		log.Println("error reseting password ", err)
		respMap["response"] = "error"
		respMap["message"] = "failed to reset user"
		return respMap
	}

	fmt.Println("reset password success")
	respMap["response"] = "success"
	return respMap
}

func ValidateToken(ctx context.Context, token string) (users.Users, bool) {
	usr := &users.Users{}

	publicKeyStr, err := os.ReadFile("public_key.pem")
	if err != nil {
		return *usr, false
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyStr)
	if err != nil {
		log.Println("error, failed to parse public key with \t err =", err)
		return *usr, false
	}

	// log.Fatalln("Login reset validateJWT")
	authentic := usr.ValidateJWT(token, publicKey)
	if !authentic {
		return *usr, false
	}
	fmt.Println("\n\t authenticated true")

	return *usr, true
}
