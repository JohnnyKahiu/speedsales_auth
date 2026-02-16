package authentication

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
	"github.com/golang-jwt/jwt/v5"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	// get post body items
	b, err := io.ReadAll(r.Body)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "could not get request params"
		return respMap
	}

	// Unmarshal into an args map
	var args map[string]string
	json.Unmarshal(b, &args)
	// if error return user not exists
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "failed to parse request params"
		return respMap
	}
	fmt.Println("username =", args["username"])

	user := users.Users{Username: args["username"]}

	// compare argon2 harshed password
	match, reset, err := user.ComparePassword(args["password"])
	if !match || err != nil {
		respMap["response"] = "error" // return success if no error and match
		respMap["message"] = "wrong username or password"
		return respMap
	}

	fmt.Println("username =", user.FirstName, " last_name =", user.LastName)
	fmt.Println("accept_payment =", user.AcceptPayment)
	fmt.Println("make_sales =", user.MakeSales)
	fmt.Println("till_num =", user.TillNum)

	privateKeyStr, err := os.ReadFile("private_key.pem")
	if err != nil {
		log.Println("error reading private key =", err)
		respMap["response"] = "error"
		respMap["message"] = "fatal error"
		return respMap
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyStr)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "could not parse private key"
		return respMap
	}

	// generate jwt token
	token, _ := user.GenerateJWT(privateKey)
	if reset {
		respMap["response"] = "reset"
		respMap["token"] = fmt.Sprintf("%v", token)
		respMap["username"] = user.Username
		respMap["message"] = "reset user password"
		return respMap
	}

	fmt.Println("generated token =", token)

	respMap["response"] = "success"
	respMap["token"] = fmt.Sprintf("%v", token)
	respMap["username"] = user.Username
	respMap["till_num"] = user.TillNum
	respMap["user_details"] = user
	return respMap
}
